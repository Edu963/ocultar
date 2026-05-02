package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/pkg/refinery"
)

// SharePointConnector implements the Connector interface for MS Graph.
type SharePointConnector struct {
	id       string
	refinery *refinery.Refinery
	client   *GraphClient

	tenantID     string
	clientID     string
	clientSecret string
	siteID       string

	deltaLink string
	stop      chan struct{}
}

func init() {
	Register("sharepoint-graph", func() Connector {
		return &SharePointConnector{
			stop: make(chan struct{}),
		}
	})
}

func (s *SharePointConnector) ID() string   { return s.id }
func (s *SharePointConnector) Type() string { return "sharepoint-graph" }

func (s *SharePointConnector) Init(config map[string]interface{}, eng *refinery.Refinery) error {
	s.id = config["id"].(string)
	s.refinery = eng

	s.tenantID = config["tenant_id"].(string)
	s.clientID = config["client_id"].(string)
	s.clientSecret = config["client_secret"].(string)

	if val, ok := config["site_id"].(string); ok {
		s.siteID = val
	}

	if s.tenantID == "" || s.clientID == "" || s.clientSecret == "" {
		return fmt.Errorf("sharepoint connector: tenant_id, client_id, and client_secret are required")
	}

	if !license.HasProConnector(license.CapProSharePoint) {
		return fmt.Errorf("sharepoint connector: requires Enterprise License with SharePoint capability enabled (Fail-Closed)")
	}

	s.client = NewGraphClient(s.tenantID, s.clientID, s.clientSecret)
	return nil
}

func (s *SharePointConnector) Start() error {
	log.Printf("[SHAREPOINT-GRAPH] Starting connector %s", s.id)
	go s.run()
	return nil
}

func (s *SharePointConnector) Stop() error {
	log.Printf("[SHAREPOINT-GRAPH] Stopping connector %s", s.id)
	close(s.stop)
	return nil
}

// Fetch implements on-demand data pull for SharePoint.
// Returns all text-processable files from the site as a JSON array of documents.
func (s *SharePointConnector) Fetch(ctx context.Context, params map[string]interface{}) ([]byte, error) {
	siteID := s.siteID
	if val, ok := params["site_id"].(string); ok && val != "" {
		siteID = val
	}
	if siteID == "" {
		return nil, fmt.Errorf("sharepoint fetch: site_id is required")
	}

	items, _, err := s.client.ListFiles(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("sharepoint fetch: %w", err)
	}

	type document struct {
		Name    string `json:"name"`
		Content string `json:"content"`
		Source  string `json:"source"`
	}

	var docs []document
	for _, item := range items {
		if item.File == nil || item.Deleted != nil || !isTextFile(item.Name) {
			continue
		}
		text, err := s.client.DownloadText(ctx, siteID, item.ID)
		if err != nil {
			log.Printf("[SHAREPOINT-GRAPH] Skipping %s: %v", item.Name, err)
			continue
		}
		docs = append(docs, document{
			Name:    item.Name,
			Content: text,
			Source:  fmt.Sprintf("sharepoint://%s/%s", siteID, item.ID),
		})
	}

	return json.Marshal(docs)
}

func (s *SharePointConnector) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.poll()
		case <-s.stop:
			return
		}
	}
}

func (s *SharePointConnector) poll() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if s.siteID == "" {
		return
	}

	log.Printf("[SHAREPOINT-GRAPH] Polling site %s (delta: %v)...", s.siteID, s.deltaLink != "")

	items, nextDelta, err := s.client.ListFiles(ctx, s.siteID)
	if err != nil {
		log.Printf("[SHAREPOINT-GRAPH] Poll error: %v", err)
		return
	}
	s.deltaLink = nextDelta

	for _, item := range items {
		if item.File == nil || item.Deleted != nil || !isTextFile(item.Name) {
			continue
		}
		text, err := s.client.DownloadText(ctx, s.siteID, item.ID)
		if err != nil {
			log.Printf("[SHAREPOINT-GRAPH] Download error for %s: %v", item.Name, err)
			continue
		}

		doc := map[string]interface{}{
			"name":    item.Name,
			"content": text,
			"source":  fmt.Sprintf("sharepoint://%s/%s", s.siteID, item.ID),
		}
		refined, err := s.refinery.ProcessInterface(doc, "sharepoint-graph-connector")
		if err != nil {
			log.Printf("[SHAREPOINT-GRAPH] Refinery error for %s: %v", item.Name, err)
			continue
		}
		log.Printf("[SHAREPOINT-REFINERY] Processed %s. PII neutralized.", item.Name)
		_ = refined
	}
}

// isTextFile returns true for file extensions the connector can process as plain text.
// Office binary formats (.docx, .xlsx, .pdf) require a text-extraction library and are skipped.
func isTextFile(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range []string{".txt", ".csv", ".json", ".md", ".log", ".xml", ".html", ".eml"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// --- GraphClient ---

// DriveItem represents a file or folder entry from the Graph API delta endpoint.
type DriveItem struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	File    *struct{} `json:"file"`    // non-nil for files, nil for folders
	Deleted *struct{} `json:"deleted"` // non-nil for deleted items
}

// GraphClient is a client for Microsoft Graph API with token caching.
type GraphClient struct {
	tenantID     string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

func NewGraphClient(tenantID, clientID, clientSecret string) *GraphClient {
	return &GraphClient{
		tenantID:     tenantID,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Authenticate performs the OAuth2 Client Credentials flow against Azure AD.
// It is a no-op if the cached token is still valid with >60s remaining.
func (c *GraphClient) Authenticate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.tokenExpiry.Add(-60 * time.Second)) {
		return nil
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", c.tenantID)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("scope", "https://graph.microsoft.com/.default")

	resp, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return fmt.Errorf("graph auth: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("graph auth: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("graph auth: parse response: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("graph auth: %s: %s", result.Error, result.ErrorDesc)
	}

	c.token = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	log.Printf("[GRAPH-API] Token refreshed, valid until %s", c.tokenExpiry.Format(time.RFC3339))
	return nil
}

func (c *GraphClient) bearerToken() (string, error) {
	if err := c.Authenticate(); err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.token, nil
}

// ListFiles fetches changed items from the site's drive using a delta query.
// On the first call deltaLink is empty and a full enumeration is performed.
// The returned deltaLink should be persisted and passed on the next call for incremental sync.
func (c *GraphClient) ListFiles(ctx context.Context, siteID string) ([]DriveItem, string, error) {
	token, err := c.bearerToken()
	if err != nil {
		return nil, "", err
	}

	// Collect all pages; Graph may paginate via @odata.nextLink.
	var items []DriveItem
	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/sites/%s/drive/root/delta", siteID)
	var finalDelta string

	for endpoint != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, "", fmt.Errorf("graph list: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("graph list: %w", err)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, "", fmt.Errorf("graph list: HTTP %d: %s", resp.StatusCode, string(body))
		}

		var page struct {
			Value     []DriveItem `json:"value"`
			NextLink  string      `json:"@odata.nextLink"`
			DeltaLink string      `json:"@odata.deltaLink"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, "", fmt.Errorf("graph list: parse: %w", err)
		}

		items = append(items, page.Value...)

		if page.DeltaLink != "" {
			finalDelta = page.DeltaLink
			endpoint = ""
		} else {
			endpoint = page.NextLink
		}
	}

	return items, finalDelta, nil
}

// DownloadText downloads a drive item's content and returns it as a string.
// The caller should verify the file is text-processable before calling this.
func (c *GraphClient) DownloadText(ctx context.Context, siteID, itemID string) (string, error) {
	token, err := c.bearerToken()
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/sites/%s/drive/items/%s/content", siteID, itemID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("graph download: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("graph download: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap per file
	if err != nil {
		return "", fmt.Errorf("graph download: read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("graph download: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
