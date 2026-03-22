package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/pkg/license"
)

// SharePointConnector implements the Connector interface for MS Graph.
type SharePointConnector struct {
	id     string
	refinery *refinery.Refinery
	client *GraphClient

	tenantID     string
	clientID     string
	clientSecret string
	siteID       string

	stop chan struct{}
}

func init() {
	Register("sharepoint-graph", func() Connector {
		return &SharePointConnector{
			stop: make(chan struct{}),
		}
	})
}

func (s *SharePointConnector) ID() string {
	return s.id
}

func (s *SharePointConnector) Type() string {
	return "sharepoint-graph"
}

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

	// Fail-Closed License Check: SharePoint requires Enterprise Tier + Bit 1 (CapProSharePoint)
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

// Fetch implements the on-demand data pull for SharePoint.
func (s *SharePointConnector) Fetch(ctx context.Context, params map[string]interface{}) ([]byte, error) {
	siteID := s.siteID
	if val, ok := params["site_id"].(string); ok {
		siteID = val
	}

	if siteID == "" {
		return nil, fmt.Errorf("sharepoint fetch: site_id is required")
	}

	log.Printf("[SHAREPOINT-GRAPH] Pulling data for site %s...", siteID)

	// Simulated pull logic matching existing simulateIngestion
	doc := map[string]interface{}{
		"name":    "HR_Onboarding_John_Doe.docx",
		"content": "Employee John Doe (john.doe@acme.com) joined on 2026-03-15. SSN: 123-45-6789.",
		"source":  fmt.Sprintf("sharepoint://%s", siteID),
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sharepoint document: %w", err)
	}

	return data, nil
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
	_, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Printf("[SHAREPOINT-GRAPH] Polling SharePoint site %s...", s.siteID)
	// In a real implementation, we would fetch files from Graph API here.
	s.simulateIngestion()
}

func (s *SharePointConnector) simulateIngestion() {
	// Refinery-First Data Flow:
	// 1. Fetch raw data via Graph API (simulated)
	documents := []map[string]interface{}{
		{
			"name":    "HR_Onboarding_John_Doe.docx",
			"content": "Employee John Doe (john.doe@acme.com) joined on 2026-03-15. SSN: 123-45-6789.",
		},
		{
			"name":    "Project_Alpha_Summary.pdf",
			"content": "Discussion with Sarah Smith regarding Project Alpha budget. Contact: +44 20 7946 0958.",
		},
	}

	for _, doc := range documents {
		// 2. Pass raw data immediately to the refinery
		refined, err := s.refinery.ProcessInterface(doc, "sharepoint-graph-connector")
		if err != nil {
			log.Printf("[SHAREPOINT-GRAPH] Error refining document %s: %v", doc["name"], err)
			continue
		}

		// 3. Only log, store, or forward the refined (tokenized) output
		log.Printf("[SHAREPOINT-REFINERY] Processed document %s. PII neutralized.", doc["name"])
		_ = refined // Refined data would be forwarded to its final destination
	}
}

// GraphClient is a simplified client for interacting with Microsoft Graph.
// In a full implementation, this might use an external SDK, but we keep it here
// or as a pkg/internal helper to avoid refinery contamination.
type GraphClient struct {
	tenantID     string
	clientID     string
	clientSecret string
	token        string
}

// NewGraphClient initializes the Graph API client wrapper.
func NewGraphClient(tenantID, clientID, clientSecret string) *GraphClient {
	return &GraphClient{
		tenantID:     tenantID,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// Authenticate performs the OAuth2 Client Credentials flow.
func (c *GraphClient) Authenticate() error {
	log.Printf("[GRAPH-API] Authenticating with Tenant %s...", c.tenantID)
	c.token = "simulated-graph-token-" + c.tenantID[:8]
	return nil
}
