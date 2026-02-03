package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ranakdinesh/citual/internal/platform/config"
	"github.com/ranakdinesh/citual/internal/platform/security"
)

func main() {
	fmt.Println("Starting Token Debugger...")

	// 1. Config
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect DB
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Fail db open: %v", err)
	}
	defer db.Close()

	// 3. Fetch Client
	clientID := "00000000-0000-0000-0000-000000000002"               // from .env / seed
	expectedSecret := "peoplesaythisisnotpossibletogoforteam0f1973%" // from .env

	var storedHash string
	err = db.QueryRow("SELECT client_secret FROM fosite_clients WHERE id=$1", clientID).Scan(&storedHash)
	if err != nil {
		log.Fatalf("Client not found in DB: %v", err)
	}
	fmt.Printf("Found Client %s.\nStored Hash: %s\n", clientID, storedHash)

	// 4. Verify Hash
	match, err := security.VerifyPassword(expectedSecret, storedHash)
	if err != nil {
		fmt.Printf("Hash Verification Error: %v\n", err)
	}
	if !match {
		fmt.Println("❌ CRITICAL: Password Verify FAILED! The DB hash does not match the secret.")

		// Attempt to see if it matches WITHOUT the % (maybe shell stripped it?)
		matchNoPercent, _ := security.VerifyPassword(strings.TrimSuffix(expectedSecret, "%"), storedHash)
		if matchNoPercent {
			fmt.Println("💡 HINT: It matches if you remove the trailing '%' !")
		} else {
			fmt.Println("   Does not match without % either.")
		}
	} else {
		fmt.Println("✅ DB Hash Verification PASSED.")
	}

	// 5. Attempt HTTP Request
	fmt.Println("\nAttempting Raw HTTP Token Exchange...")

	// Create request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", "DUMMY_CODE_JUST_TESTING_AUTH") // Invalid code, but we want to see if we get 'invalid_grant' (auth success) vs 'invalid_request' (auth fail)
	data.Set("redirect_uri", "http://localhost:3000/api/auth/callback")

	req, _ := http.NewRequest("POST", "http://localhost:8090/oauth2/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + expectedSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("HTTP Req failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))

	if resp.StatusCode == 400 {
		if strings.Contains(string(body), "invalid_client") || strings.Contains(string(body), "invalid_request") {
			fmt.Println("❌ Client Auth still failing on server.")
		} else if strings.Contains(string(body), "invalid_grant") {
			fmt.Println("✅ Client Auth SUCCEEDED! (We got invalid_grant as expected for dummy code).")
		}
	}
}
