package identities

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/license"
)

// StartSyncWorker boots a background goroutine that polls CRM/LDAP APIs
// and synchronizes fetched identities into the config's Tier 0 Dictionary.
func StartSyncWorker() {
	if !license.IsEnterprise() {
		return
	}

	interval, err := time.ParseDuration(config.Global.SyncInterval)
	if err != nil {
		log.Printf("[ERROR] CRM Sync: Invalid sync_interval '%s', defaulting to 5m", config.Global.SyncInterval)
		interval = 5 * time.Minute
	}

	go func() {
		log.Printf("[INFO] Live CRM Identity Sync Worker started (polling %s every %v)", config.Global.CRMEndpoint, interval)
		
		// Initial sync
		performSync()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			performSync()
		}
	}()
}

func performSync() {
	endpoint := config.Global.CRMEndpoint
	if endpoint == "" {
		log.Println("[DEBUG] CRM Sync: skipping poll, crm_endpoint not configured.")
		return
	}

	apiKey := config.Global.CRMApiKey
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Printf("[ERROR] CRM Sync: failed to create request: %v", err)
		return
	}
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] CRM Sync: polling failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] CRM Sync: unexpected status code: %d", resp.StatusCode)
		return
	}

	var identities []string
	if err := json.NewDecoder(resp.Body).Decode(&identities); err != nil {
		log.Printf("[ERROR] CRM Sync: failed to decode response: %v", err)
		return
	}

	added := 0
	for _, id := range identities {
		config.AddDictionaryTerm("PROTECTED_ENTITY", id)
		added++
	}

	if added > 0 {
		if err := config.Save(); err != nil {
			log.Printf("[ERROR] CRM Sync: failed to save synced identities: %v", err)
		} else {
			log.Printf("[INFO] CRM Sync: Automatically ingested %d protected identities.", added)
		}
	}
}
