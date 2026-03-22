// Package proxy implements the OCULTAR transparent HTTP reverse-proxy handler.
// It performs request-time PII redaction and response-time token re-hydration.
package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

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
}

// NewHandler constructs a Handler pointed at the given upstream targetURL.
// The refinery is used for Tier 1 + Tier 2 redaction on the request body.
// masterKey and vault are used for re-hydration on the response body.
func NewHandler(eng *refinery.Refinery, v vault.Provider, masterKey []byte, targetURL string) (*Handler, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid OCU_PROXY_TARGET %q: %w", targetURL, err)
	}
	return &Handler{
		eng:       eng,
		vault:     v,
		masterKey: masterKey,
		target:    u,
		transport: &http.Transport{
			// Disable auto-decompression so we handle the body ourselves.
			DisableCompression: true,
			// Sensible timeouts for production traffic.
			ResponseHeaderTimeout: 120 * time.Second,
		},
		sem: make(chan struct{}, 15), // match postgres connection limit
	}, nil
}

// ServeHTTP implements http.Handler. It:
//  1. Reads the incoming request body.
//  2. If the body is JSON, runs it through the OCULTAR refinery (PII → tokens).
//  3. Forwards the sanitised request to the upstream target.
//  4. Reads the upstream response body.
//  5. If the response is JSON, scans for vault tokens and re-hydrates them.
//  6. Returns the final response to the original caller.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Rate limiting / concurrency control
	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
		// Performance Shield: If capacity > 85%, skip heavy Tier 2 AI to prevent latency thrashing.
		load := float64(len(h.sem)) / float64(cap(h.sem))
		if load > 0.85 {
			h.eng.SkipDeepScan = true
			log.Printf("[PERF-SHIELD] High load detected (%.2f%%). Bypassing Tier 2 Deep Scan.", load*100)
		} else {
			h.eng.SkipDeepScan = false
		}
	case <-time.After(5 * time.Second):
		http.Error(w, "ocultar-proxy: too many requests (concurrency limit reached)", http.StatusTooManyRequests)
		return
	}

	actor := r.Header.Get("X-Forwarded-For")
	if actor == "" {
		actor = r.RemoteAddr
	}

	// ── 1. Read incoming body ────────────────────────────────────────────────
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024) // 5MB limit
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "ocultar-proxy: payload too large", http.StatusRequestEntityTooLarge)
		return
	}
	defer r.Body.Close()

	// ── 2. Redact PII from JSON body ─────────────────────────────────────────
	sanitisedBody, redacted, err := h.redactBody(rawBody, actor)
	if err != nil {
		log.Printf("[PROXY-BLOCK] Refinery error: %v", err)
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

	// Copy request headers (skip hop-by-hop and our custom headers).
	copyRequestHeaders(upstreamReq.Header, r.Header)
	upstreamReq.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	if redacted {
		upstreamReq.Header.Set(headerRedacted, "true")
	}
	upstreamReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(sanitisedBody)))

	log.Printf("[PROXY] → %s %s (body: %d bytes, redacted: %v)", r.Method, upstreamURL, len(sanitisedBody), redacted)

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
	finalBody, err := h.rehydrateBody(respBody, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("[PROXY] re-hydration failed: %v", err)
		http.Error(w, "ocultar-proxy: re-hydration failed (data loss protection)", http.StatusInternalServerError)
		return
	}

	log.Printf("[PROXY] ← %d  (response: %d bytes)", resp.StatusCode, len(finalBody))

	// ── 7. Write response back to the client ──────────────────────────────────
	copyResponseHeaders(w.Header(), resp.Header)
	// Adjust Content-Length if the body length changed after re-hydration.
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(finalBody)))
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(finalBody); err != nil {
		log.Printf("[PROXY] error writing response to client: %v", err)
	}
}

// redactBody parses the body as JSON and runs it through the refinery's full
// pipeline (Tier 0 → Tier 2). Returns the redacted body bytes, a boolean
// indicating whether any PII was found and replaced, and any critical refinery error.
// If the body is not valid JSON it is passed through the refinery line-by-line as
// plain text, preserving the original Content-Type behaviour.
func (h *Handler) redactBody(body []byte, actor string) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}

	// Obfuscation detection
	bodyStr := string(body)
	if strings.Contains(bodyStr, "%7B") && strings.Contains(bodyStr, "%22") {
		return nil, false, fmt.Errorf("obfuscated payload detected: url-encoded JSON")
	}
	if strings.HasPrefix(strings.TrimSpace(bodyStr), "ey") && !strings.Contains(bodyStr, " ") && len(bodyStr) > 50 {
		return nil, false, fmt.Errorf("obfuscated payload detected: base64/JWT")
	}

	h.eng.ResetHits()

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var outBuf bytes.Buffer

	// Try JSON streaming
	if err := streamRefineJSON(dec, h.eng, actor, &outBuf); err == nil {
		if _, err := dec.Token(); err == io.EOF {
			report := h.eng.GenerateReport(1)
			return outBuf.Bytes(), report.TotalCount > 0, nil
		}
	}

	// Plain-text / non-JSON path: process line by line.
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

// rehydrateBody scans the response body for vault tokens and replaces them
// with the original PII values from the vault provider.
// Non-JSON responses are scanned as raw text (tokens are plain ASCII so this
// is safe for any Content-Type).
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

	// Fallback: raw byte scan with the token regex.
	res, err := RehydrateString(h.vault, h.masterKey, string(body))
	if err != nil {
		return nil, err
	}
	return []byte(res), nil
}

// resolveTarget builds the full upstream URL from the incoming request.
// A per-request override is accepted via the Ocultar-Target header,
// but is validated against SSRF attacks.
func (h *Handler) resolveTarget(r *http.Request) (string, error) {
	base := h.target.String()
	if override := r.Header.Get(headerTarget); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("invalid Ocultar-Target URL")
		}
		host := parsed.Hostname()
		if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "169.254.") {
			return "", fmt.Errorf("SSRF blocked: internal target '%s' not allowed", host)
		}
		base = strings.TrimRight(override, "/")
	}
	path := r.URL.RequestURI() // includes query string
	return base + path, nil
}

// ── Header helpers ────────────────────────────────────────────────────────────

// hopByHopHeaders are headers that must NOT be forwarded between proxy hops.
var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
	// Our custom headers are consumed by the proxy itself.
	headerTarget: true,
}

func copyRequestHeaders(dst, src http.Header) {
	for k, vv := range src {
		if hopByHopHeaders[k] {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyResponseHeaders(dst, src http.Header) {
	for k, vv := range src {
		if hopByHopHeaders[k] {
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
		if v == '{' {
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
		} else if v == '[' {
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
		if val == '{' {
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
		} else if val == '[' {
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
