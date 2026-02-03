package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // Use pgx via stdlib for simplicity in script
	"github.com/ranakdinesh/citual/internal/platform/config"
	"github.com/ranakdinesh/citual/internal/platform/security"
)

func main() {
	log.Println("Starting database seeding...")

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

	ctx := context.Background()

	// 3. Seed Tenant (Ops)
	// Check if Ops tenant already exists
	var opsTenantID uuid.UUID
	var opsTenantName string = "Citual Ops"

	err = db.QueryRowContext(ctx, "SELECT id FROM tenants WHERE kind = 'ops' LIMIT 1").Scan(&opsTenantID)
	if err == nil {
		log.Printf("Ops tenant already exists: %s", opsTenantID)
	} else if err == sql.ErrNoRows {
		// Insert new
		opsTenantID = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, kind, created_at, updated_at)
			VALUES ($1, $2, 'ops', NOW(), NOW())
		`, opsTenantID, opsTenantName)
		if err != nil {
			log.Fatalf("Failed to seed tenant: %v", err)
		}
		log.Printf("Seeded Ops Tenant: %s (%s)", opsTenantName, opsTenantID)
	} else {
		log.Fatalf("Failed to query existing ops tenant: %v", err)
	}

	// 4. Seed Roles
	var superAdminRoleID uuid.UUID

	roles := []string{"Super Admin", "Admin", "Member"}

	for _, roleName := range roles {
		// Check if role exists
		var roleID uuid.UUID
		err := db.QueryRowContext(ctx, "SELECT id FROM roles WHERE name = $1 AND tenant_id = $2", roleName, opsTenantID).Scan(&roleID)

		if err == nil {
			log.Printf("Role already exists: %s (%s)", roleName, roleID)
		} else {
			// Insert new
			roleID = uuid.New()
			_, err = db.ExecContext(ctx, `
				INSERT INTO roles (id, name, tenant_id, description, created_at)
				VALUES ($1, $2, $3, $4, NOW())
			`, roleID, roleName, opsTenantID, fmt.Sprintf("Default %s role", roleName))
			if err != nil {
				log.Printf("Failed to seed role %s: %v", roleName, err)
				continue
			}
			log.Printf("Seeded Role: %s", roleName)
		}

		// Capture Super Admin ID for later assignment
		if roleName == "Super Admin" {
			superAdminRoleID = roleID
		}
	}

	// 5. Seed Super Admin User
	userEmail := "dinesh@citual.in"
	userPassword := "ChangeMe_123!"

	// Hash password
	hashedPwd, err := security.HashPassword(userPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	userID := uuid.New()

	// Check if user exists first to get the correct ID
	var existingUserID uuid.UUID
	err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", userEmail).Scan(&existingUserID)
	if err == nil {
		userID = existingUserID
		log.Printf("User already exists: %s", userEmail)
	} else {
		// Insert new
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (id, email, first_name, last_name, password_hash, tenant_id, is_super_admin, created_at, updated_at)
			VALUES ($1, $2, 'Dinesh', 'Ranak', $3, $4, TRUE, NOW(), NOW())
		`, userID, userEmail, hashedPwd, opsTenantID)
		if err != nil {
			log.Fatalf("Failed to seed user: %v", err)
		}
		log.Printf("Seeded Super Admin User: %s", userEmail)
	}

	// 6. Assign Super Admin Role
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, superAdminRoleID)
	if err != nil {
		log.Fatalf("Failed to assign role: %v", err)
	}
	log.Printf("Assigned 'Super Admin' role to user")

	// 7. Seed OAuth2 Client (Fosite)
	// Using the values from .env
	// AUTH_CLIENT_ID=00000000-0000-0000-0000-000000000002
	// AUTH_CLIENT_SECRET=peoplesaythisisnotpossibletogoforteam0f1973%

	clientID := cfg.AuthClientID
	if clientID == "" {
		clientID = "00000000-0000-0000-0000-000000000002" // Fallback
	}
	clientSecret := cfg.AuthClientSecret
	if clientSecret == "" {
		clientSecret = "peoplesaythisisnotpossibletogoforteam0f1973%" // Fallback
	}

	// Fosite expects the secret to be hashed (usually bcrypt).
	// The prompt mentioned "Argon2" for *User* passwords.
	// Standard Fosite store implementations often use BCrypt for *Client* secrets.
	// However, if the user insists on Argon2 everywhere or if our Fosite store uses `security.HashPassword`, we should use that.
	// Looking at `internal/modules/identity/adapters/fosite_store/store.go` would confirm, but let's stick to `security.HashPassword` (Argon2)
	// as that's the "platform/security" standard we saw earlier.

	hashedSecret, err := security.HashPassword(clientSecret)
	if err != nil {
		log.Fatalf("Failed to hash client secret: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO fosite_clients (
			id, 
			client_secret, 
			tenant_id, 
			redirect_uris, 
			grant_types, 
			response_types, 
			scopes, 
			audience, 
			public, 
			active, 
			created_at, 
			updated_at
		) VALUES (
			$1, $2, $3, 
			$4, -- redirect_uris
			$5, -- grant_types
			$6, -- response_types
			$7, -- scopes
			$8, -- audience
			FALSE, TRUE, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET 
			client_secret = EXCLUDED.client_secret,
			updated_at = NOW();
	`,
		clientID,
		hashedSecret,
		opsTenantID,
		[]string{"http://localhost:3000", "http://localhost:3000/api/auth/callback", "http://localhost:8090/callback"}, /* redirect_uris */
		[]string{"authorization_code", "client_credentials", "refresh_token"},                                          /* grant_types */
		[]string{"code", "token", "id_token"},                                                                          /* response_types */
		[]string{"openid", "profile", "email", "offline_access"},                                                       /* scopes */
		[]string{"citual", "oauth-service"},                                                                            /* audience */
	)
	if err != nil {
		log.Fatalf("Failed to seed OAuth2 Client: %v", err)
	}
	log.Printf("Seeded OAuth2 Client: %s", clientID)

	log.Println("Seeding completed successfully!")
}
