// Package proxy implements the OCULTAR transparent HTTP reverse-proxy handler.
// It performs request-time PII redaction and response-time token re-hydration.
package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
)

const (
	// headerRedacted is added to upstream requests so operators can audit that
	// the proxy processed the request.
	headerRedacted = "X-Ocultar-Redacted"
	// headerTarget allows per-request override of the upstream target URL.
	headerTarget = "Ocultar-Target"
)

// Handler is the OCULTAR proxy HTTP handler.
// It holds references to the shared refinery and vault for thread-safe operation.
type Handler struct {
	eng       *refinery.Refinery
	vault     vault.Provider
	masterKey []byte
	target    *url.URL
	transport http.RoundTripper
	sem       chan struct{}
	waitQueue chan struct{}
}

// NewHandler constructs a Handler pointed at the given upstream targetURL.
// The refinery is used for Tier 1 + Tier 2 redaction on the request body.
// masterKey and vault are used for re-hydration on the response body.
func NewHandler(eng *refinery.Refinery, v vault.Provider, masterKey []byte, targetURL string) (*Handler, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid OCU_PROXY_TARGET %q: %w", targetURL, err)
	}

	concurrency := config.Global.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 15 // Legacy default
	}

	queueSize := config.Global.QueueSize
	if queueSize <= 0 {
		queueSize = 100
	}

	return &Handler{
		eng:       eng,
		vault:     v,
		masterKey: masterKey,
		target:    u,
		transport: &http.Transport{
			DisableCompression: true,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: 120 * time.Second,
		},
		sem:       make(chan struct{}, concurrency),
		waitQueue: make(chan struct{}, queueSize),
	}, nil
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// ── 0. Concurrency & Queueing (Fail-Closed) ─────────────────────────────
	select {
	case h.waitQueue <- struct{}{}:
		QueueLength.Inc()
		defer func() {
			<-h.waitQueue
			QueueLength.Dec()
		}()
	default:
		DroppedRequestsTotal.WithLabelValues("queue_full").Inc()
		http.Error(w, "ocultar-proxy: too many requests (queue full)", http.StatusTooManyRequests)
		return
	}

	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	case <-time.After(10 * time.Second): // Wait up to 10s for a worker
		DroppedRequestsTotal.WithLabelValues("timeout").Inc()
		http.Error(w, "ocultar-proxy: internal timeout waiting for refinery worker", http.StatusServiceUnavailable)
		return
	case <-r.Context().Done():
		return
	}

	actor := r.Header.Get("X-Forwarded-For")
	if actor == "" {
		actor = r.RemoteAddr
	}

	// ── 1. Read incoming body ────────────────────────────────────────────────
	maxSize := config.Global.MaxPayloadSize
	if maxSize <= 0 {
		maxSize = 5 * 1024 * 1024
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		DroppedRequestsTotal.WithLabelValues("payload_too_large").Inc()
		http.Error(w, "ocultar-proxy: payload exceeded configured limit", http.StatusRequestEntityTooLarge)
		return
	}
	defer r.Body.Close()

	// ── 2. Redact PII from JSON body ─────────────────────────────────────────
	refStart := time.Now()
	sanitisedBody, redacted, err := h.redactBody(rawBody, actor)
	log.Printf("[DEBUG] Redaction complete. Redacted: %v, Body length: %d -> %d", redacted, len(rawBody), len(sanitisedBody))
	if redacted {
		log.Printf("[DEBUG] Sanitised Body (truncated): %s", string(sanitisedBody))
	}
	RequestLatency.WithLabelValues("refinery_total").Observe(time.Since(refStart).Seconds())

	if err != nil {
		log.Printf("[PROXY-BLOCK] Refinery error: %v", err)
		RequestsTotal.WithLabelValues(r.Method, "error", "false").Inc()
		if strings.Contains(err.Error(), "trial limit reached") {
			http.Error(w, "ocultar-proxy: trial limit reached (fail-closed)", http.StatusForbidden)
		} else {
			http.Error(w, "ocultar-proxy: internal security refinery failure (fail-closed)", http.StatusInternalServerError)
		}
		return
	}

	// ── 3. Construct the upstream request ────────────────────────────────────
	upstreamURL, err := h.resolveTarget(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("ocultar-proxy: %v", err), http.StatusForbidden)
		return
	}
	upstreamReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(sanitisedBody))
	if err != nil {
		http.Error(w, "ocultar-proxy: failed to build upstream request", http.StatusInternalServerError)
		return
	}

	copyRequestHeaders(upstreamReq.Header, r.Header)
	upstreamReq.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	if redacted {
		upstreamReq.Header.Set(headerRedacted, "true")
	}
	upstreamReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(sanitisedBody)))

	// ── 4. Forward to upstream ───────────────────────────────────────────────
	resp, err := h.transport.RoundTrip(upstreamReq)
	if err != nil {
		log.Printf("[PROXY] upstream error: %v", err)
		http.Error(w, fmt.Sprintf("ocultar-proxy: upstream error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// ── 5. Read upstream response ─────────────────────────────────────────────
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "ocultar-proxy: failed to read upstream response", http.StatusBadGateway)
		return
	}

	// ── 6. Re-hydrate tokens in response ─────────────────────────────────────
	rehyStart := time.Now()
	finalBody, err := h.rehydrateBody(respBody, resp.Header.Get("Content-Type"))
	RequestLatency.WithLabelValues("rehydration").Observe(time.Since(rehyStart).Seconds())

	if err != nil {
		log.Printf("[PROXY] re-hydration failed: %v", err)
		if config.Global.RehydrateFallbackEnabled {
			log.Printf("[WARN] Re-hydration failed, falling back to tokenized response (Safety: ON)")
			finalBody = respBody // Return tokens instead of leaking data or failing
		} else {
			http.Error(w, "ocultar-proxy: re-hydration failed (strict data loss protection)", http.StatusInternalServerError)
			return
		}
	}

	// ── 7. Write response back to the client ──────────────────────────────────
	copyResponseHeaders(w.Header(), resp.Header)
	if redacted {
		w.Header().Set(headerRedacted, "true")
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(finalBody)))
	w.WriteHeader(resp.StatusCode)
	w.Write(finalBody)

	statusStr := fmt.Sprintf("%d", resp.StatusCode)
	redactedStr := fmt.Sprintf("%v", redacted)
	RequestsTotal.WithLabelValues(r.Method, statusStr, redactedStr).Inc()
	RequestLatency.WithLabelValues("total").Observe(time.Since(start).Seconds())
}

func (h *Handler) redactBody(body []byte, actor string) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, "%7B") && strings.Contains(bodyStr, "%22") {
		return nil, false, fmt.Errorf("obfuscated payload detected: url-encoded JSON")
	}
	if (strings.HasPrefix(strings.TrimSpace(bodyStr), "ey") || strings.HasPrefix(strings.TrimSpace(bodyStr), "eyJ")) && 
		!strings.Contains(bodyStr, " ") && len(bodyStr) > 50 {
		return nil, false, fmt.Errorf("obfuscated payload detected: base64/JWT")
	}

	h.eng.ResetHits()

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var outBuf bytes.Buffer

	if err := streamRefineJSON(dec, h.eng, actor, &outBuf); err == nil {
		if _, err := dec.Token(); err == io.EOF {
			report := h.eng.GenerateReport(1)
			return outBuf.Bytes(), report.TotalCount > 0, nil
		}
	}

	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			refined, refErr := h.eng.RefineString(line, actor, nil)
			if refErr != nil {
				return nil, false, refErr
			}
			lines[i] = refined
		}
	}
	report := h.eng.GenerateReport(1)
	return []byte(strings.Join(lines, "\n")), report.TotalCount > 0, nil
}

func (h *Handler) rehydrateBody(body []byte, contentType string) ([]byte, error) {
	if len(body) == 0 || !ContainsTokensInBody(body) {
		return body, nil
	}

	isJSON := strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/json")

	if isJSON {
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()
		var outBuf bytes.Buffer
		if err := streamRehydrateJSON(dec, h.vault, h.masterKey, &outBuf); err == nil {
			if _, err := dec.Token(); err == io.EOF {
				return outBuf.Bytes(), nil
			}
		}
	}

	res, err := RehydrateString(h.vault, h.masterKey, string(body))
	if err != nil {
		return nil, err
	}
	return []byte(res), nil
}

// resolveTarget builds the full upstream URL from the incoming request.
// It includes full RFC 1918, IPv6, and loopback protection with DNS rebinding safety.
func (h *Handler) resolveTarget(r *http.Request) (string, error) {
	base := h.target.String()
	targetHeader := r.Header.Get(headerTarget)
	if targetHeader == "" {
		return base + r.URL.RequestURI(), nil
	}

	parsed, err := url.Parse(targetHeader)
	if err != nil {
		return "", fmt.Errorf("invalid Ocultar-Target URL")
	}

	// Hostname() correctly strips IPv6 brackets (e.g. "[::1]" → "::1")
	// so that net.ParseIP and net.LookupIP receive a bare address or name.
	host := parsed.Hostname()

	// DNS Rebinding Protection: Resolve names to IPs and validate IPs
	ips, err := net.LookupIP(host)
	if err != nil {
		// If it's already an IP, it will be validated by isPrivateIP
		if ip := net.ParseIP(host); ip != nil {
			ips = []net.IP{ip}
		} else {
			return "", fmt.Errorf("failed to resolve target host: %v", err)
		}
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return "", fmt.Errorf("SSRF blocked: internal/private target '%s' (%s) not allowed", host, ip.String())
		}
	}

	return strings.TrimRight(targetHeader, "/") + r.URL.RequestURI(), nil
}

func isPrivateIP(ip net.IP) bool {
	// Unspecified addresses (0.0.0.0, ::) must be blocked — they can be used
	// to bypass address checks while still reaching the local host on many OSes.
	if ip.IsUnspecified() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}

	// RFC 1918 (IPv4) & RFC 4193 (IPv6 Unique Local Address)
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}

	// IPv6 ULA (fc00::/7)
	return len(ip) == net.IPv6len && (ip[0]&0xfe) == 0xfc
}

var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
	"Content-Length":      true,
	headerTarget:          true,
}

func copyRequestHeaders(dst, src http.Header) {
	for k, vv := range src {
		if hopByHopHeaders[strings.Title(k)] || hopByHopHeaders[k] {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyResponseHeaders(dst, src http.Header) {
	for k, vv := range src {
		if hopByHopHeaders[strings.Title(k)] || hopByHopHeaders[k] {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func streamRefineJSON(dec *json.Decoder, eng *refinery.Refinery, actor string, out *bytes.Buffer) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := t.(type) {
	case json.Delim:
		out.WriteString(v.String())
		switch v {
		case '{':
			first := true
			for dec.More() {
				if !first {
					out.WriteString(",")
				}
				first = false
				kt, err := dec.Token()
				if err != nil {
					return err
				}
				b, _ := json.Marshal(kt)
				out.Write(b)
				out.WriteString(":")
				if err := streamRefineJSON(dec, eng, actor, out); err != nil {
					return err
				}
			}
			et, err := dec.Token()
			if err != nil {
				return err
			}
			out.WriteString(et.(json.Delim).String())
		case '[':
			first := true
			for dec.More() {
				if !first {
					out.WriteString(",")
				}
				first = false
				if err := streamRefineJSON(dec, eng, actor, out); err != nil {
					return err
				}
			}
			et, err := dec.Token()
			if err != nil {
				return err
			}
			out.WriteString(et.(json.Delim).String())
		}
	case string:
		sanitised, err := eng.ProcessInterface(v, actor)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(sanitised)
		out.Write(b)
	default:
		b, _ := json.Marshal(v)
		out.Write(b)
	}
	return nil
}

func streamRehydrateJSON(dec *json.Decoder, v vault.Provider, masterKey []byte, out *bytes.Buffer) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	switch val := t.(type) {
	case json.Delim:
		out.WriteString(val.String())
		switch val {
		case '{':
			first := true
			for dec.More() {
				if !first {
					out.WriteString(",")
				}
				first = false
				kt, err := dec.Token()
				if err != nil {
					return err
				}
				b, _ := json.Marshal(kt)
				out.Write(b)
				out.WriteString(":")
				if err := streamRehydrateJSON(dec, v, masterKey, out); err != nil {
					return err
				}
			}
			et, err := dec.Token()
			if err != nil {
				return err
			}
			out.WriteString(et.(json.Delim).String())
		case '[':
			first := true
			for dec.More() {
				if !first {
					out.WriteString(",")
				}
				first = false
				if err := streamRehydrateJSON(dec, v, masterKey, out); err != nil {
					return err
				}
			}
			et, err := dec.Token()
			if err != nil {
				return err
			}
			out.WriteString(et.(json.Delim).String())
		}
	case string:
		hydrated, err := RehydrateString(v, masterKey, val)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(hydrated)
		out.Write(b)
	default:
		b, _ := json.Marshal(val)
		out.Write(b)
	}
	return nil
}
