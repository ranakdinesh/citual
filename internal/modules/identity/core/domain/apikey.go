package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Prefix      string
	KeyHash     string // Store hash only!
	Scopes      []string
	IPAllowlist []string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}
