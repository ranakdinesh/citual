package domain

import (
	"github.com/google/uuid"
	"time"
)

type VerificationKind string

const (
	VerificationKindEmailVerify  VerificationKind = "email_verify"
	VerificationKindPassReset    VerificationKind = "password_reset"
	VerificationKindMobileVerify VerificationKind = "mobile_verify"
)

type VerificationChallenge struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Kind       VerificationKind
	TokenHash  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

func NewChallenge(tenantID, userID uuid.UUID, kind VerificationKind, tokenHash string, duration time.Duration) *VerificationChallenge {
	return &VerificationChallenge{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		Kind:      kind,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(duration),
		CreatedAt: time.Now().UTC(),
	}
}
