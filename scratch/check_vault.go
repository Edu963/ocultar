package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	db, err := sql.Open("duckdb", "vault.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM vault").Scan(&count)
	if err != nil {
		// Table might not exist yet
		fmt.Printf("Vault table count: 0 (or error: %v)\n", err)
		return
	}
	fmt.Printf("Vault contains %d tokens.\n", count)
}
