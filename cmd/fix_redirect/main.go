package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq" // Using pq for array support helper if needed, or just manual string building, but pgx handles arrays well natively. Actually standard sql/pgx driver handles string arrays as {a,b}.
	"github.com/ranakdinesh/citual/internal/platform/config"
)

func main() {
	log.Println("Starting redirect URI fix...")

	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	clientID := "00000000-0000-0000-0000-000000000002"
	newURI := "http://localhost:3000/api/auth/callback"

	// Fetch current URIs
	var uris []string
	err = db.QueryRow("SELECT redirect_uris FROM fosite_clients WHERE id = $1", clientID).Scan(pq.Array(&uris))
	if err != nil {
		log.Fatalf("Failed to fetch client: %v", err)
	}

	// Check if already exists
	exists := false
	for _, u := range uris {
		if u == newURI {
			exists = true
			break
		}
	}

	if exists {
		log.Println("URI already exists, skipping.")
		return
	}

	// Add and update
	uris = append(uris, newURI)
	_, err = db.Exec("UPDATE fosite_clients SET redirect_uris = $1 WHERE id = $2", pq.Array(uris), clientID)
	if err != nil {
		log.Fatalf("Failed to update client: %v", err)
	}

	log.Printf("Successfully added %s to client %s", newURI, clientID)
}
