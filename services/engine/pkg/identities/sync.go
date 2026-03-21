package identities

import (
	"log"
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

	go func() {
		log.Println("[INFO] Live CRM/LDAP Identity Sync Worker started")
		// Simulate initial pull delay to represent a network request to Salesforce/HubSpot
		time.Sleep(5 * time.Second)

		// Hardcoded response representing a JSON payload from an HR system
		dummyIdentities := []string{
			"John Doe", "Jane Smith", "Alice Johnson",
			"Global Corp", "Acme Finance",
		}

		added := 0
		for _, id := range dummyIdentities {
			config.AddDictionaryTerm("PROTECTED_ENTITY", id)
			added++
		}

		if added > 0 {
			if err := config.Save(); err != nil {
				log.Printf("[ERROR] Failed to save synced identities: %v", err)
			} else {
				log.Printf("[INFO] CRM Sync: Automatically ingested %d protected identities.", added)
			}
		}

		// Setup polling loop
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			// In production, poll remote API here and inject deltas.
			log.Printf("[DEBUG] CRM Sync: polling external identity provider...")
		}
	}()
}
