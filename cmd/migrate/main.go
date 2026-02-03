package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ranakdinesh/citual/internal/platform/config"
)

func main() {
	log.Println("Starting database migrations...")

	// 1. Load Config
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to DB
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// 3. Read Migration Files
	migrationDir := "internal/modules/identity/sql/migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatalf("Failed to read migration directory: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles) // Ensure order: 0001, 0002, ...

	// 4. Execute Migrations
	for _, filename := range sqlFiles {
		log.Printf("Applying migration: %s", filename)
		content, err := os.ReadFile(filepath.Join(migrationDir, filename))
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Failed to apply migration %s: %v", filename, err)
		}
	}

	log.Println("All migrations applied successfully!")
}
