//go:build ignore

package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/vault"

	_ "github.com/lib/pq"
)

const (
	dsn         = "postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
	numNodes    = 5
	concurrency = 50
	numRecords  = 1000
)

func generateRandomHash() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	log.Println("[INFO] Starting HA Cluster Verification & Benchmark")

	nodes := make([]vault.Provider, numNodes)
	var nodesMu sync.RWMutex
	var err error

	// Truncate the table at start to ensure clean benchmark metrics
	db, err := sql.Open("postgres", dsn)
	if err == nil {
		db.Exec("TRUNCATE TABLE vault;")
		db.Close()
	}

	for i := 0; i < numNodes; i++ {
		nodes[i], err = vault.New(config.Settings{VaultBackend: "postgres", PostgresDSN: dsn}, "")
		if err != nil {
			log.Fatalf("Failed to initialize node %d: %v", i, err)
		}
		log.Printf("[INFO] Node %d initialized successfully.", i)
	}

	// Prepare records
	log.Printf("[INFO] Generating %d unique PII hashes...", numRecords)
	records := make([]string, numRecords)
	for i := 0; i < numRecords; i++ {
		records[i] = generateRandomHash()
	}

	start := time.Now()
	var wg sync.WaitGroup

	log.Printf("[INFO] Launching %d concurrent workers across %d nodes...", concurrency, numNodes)

	// Benchmark Concurrent Writes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			nodeIdx := workerID % numNodes

			for j := 0; j < numRecords; j++ {
				hash := records[j]
				token := fmt.Sprintf("[TOKEN_%s]", hash[:8])
				encPII := "ENCRYPTED_" + hash

				nodesMu.RLock()
				node := nodes[nodeIdx]
				nodesMu.RUnlock()

				// If node is nil or closed, we will naturally error here and log it.
				// But we continue to simulate an external retry logic.
				if node == nil {
					continue
				}

				// In a real cluster, nodes race to insert. Safe due to ON CONFLICT DO NOTHING.
				_, err := node.StoreToken(hash, token, encPII)
				if err != nil {
					log.Printf("[ERROR] Worker %d (Node %d) failed insertion: %v", workerID, nodeIdx, err)
					time.Sleep(10 * time.Millisecond) // backoff on failure
				}
			}
		}(i)
	}

	// Simulate node failure and rolling restart
	go func() {
		time.Sleep(1 * time.Second)
		log.Println("[WARN] SIMULATING NODE FAILURE: Dropping cluster node 1...")

		nodesMu.Lock()
		nodes[1].Close()
		nodes[1] = nil
		nodesMu.Unlock()

		time.Sleep(500 * time.Millisecond)
		log.Println("[INFO] SIMULATING ROLLING RESTART: Recovering cluster node 1...")

		newNode, err := vault.New(config.Settings{VaultBackend: "postgres", PostgresDSN: dsn}, "")

		nodesMu.Lock()
		nodes[1] = newNode
		nodesMu.Unlock()

		if err != nil {
			log.Printf("[ERROR] Rolling restart failed for node 1: %v", err)
		} else {
			log.Println("[INFO] Node 1 recovered gracefully.")
		}
	}()

	wg.Wait()
	duration := time.Since(start)

	log.Println("========================================")
	log.Println("           HA BENCHMARK RESULTS         ")
	log.Println("========================================")
	log.Printf("Total duration    : %v", duration)
	log.Printf("Ops per second    : %.0f inserts/sec", float64(numRecords*concurrency)/duration.Seconds())

	// Clean shutdown check
	for i, n := range nodes {
		if n == nil {
			log.Printf("[WARN] Node %d is nil, skipping close.", i)
			continue
		}
		if err := n.Close(); err != nil {
			log.Printf("[WARN] Node %d failed to cleanly shut down: %v", i, err)
		} else {
			log.Printf("[INFO] Node %d gracefully detached from cluster.", i)
		}
	}

	// Final validation check with a fresh connection
	validator, _ := vault.New(config.Settings{VaultBackend: "postgres", PostgresDSN: dsn}, "")
	defer validator.Close()

	count := validator.CountAll()
	if count == int64(numRecords) {
		log.Printf("[SUCCESS] Vault Idempotency Verified! Expected %d records, got %d.", numRecords, count)
	} else {
		log.Fatalf("[FAILED] Cluster race condition detected! Expected %d records, got %d.", numRecords, count)
		os.Exit(1)
	}

	log.Println("[SUCCESS] Postgres HA stateless scaling validated successfully in multi-node topology.")
}
