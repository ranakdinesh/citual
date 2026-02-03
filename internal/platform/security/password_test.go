package security

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "SecretPassword123!"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if hash == "" {
		t.Fatal("Hash is empty")
	}
	
	ok, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !ok {
		t.Error("Password verification failed for correct password")
	}
	
	ok, err = VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("Error during verification: %v", err)
	}
	if ok {
		t.Error("Password verification succeeded for wrong password")
	}
}
