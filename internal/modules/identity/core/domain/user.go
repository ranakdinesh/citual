package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidFirstName = errors.New("invalid first name")
	ErrInvalidLastName  = errors.New("invalid last name")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidMobile    = errors.New("invalid mobile")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUserInactive     = errors.New("user is inactive")
)

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	FirstName    string
	LastName     string
	Mobile       string // Added
	Email        string
	PasswordHash string
	IsSuperAdmin bool
	AuthzVersion int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Verification Status
	EmailVerifiedAt  *time.Time
	MobileVerifiedAt *time.Time
}

// NewUser Factory
func NewUser(tenantID uuid.UUID, firstName, lastName, email, mobile string) (*User, error) {
	if strings.TrimSpace(firstName) == "" {
		return nil, ErrInvalidFirstName
	}
	if strings.TrimSpace(lastName) == "" {
		return nil, ErrInvalidLastName
	}
	if !looksLikeEmail(email) {
		return nil, ErrInvalidEmail
	}

	return &User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     strings.ToLower(email),
		Mobile:    mobile,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// Rich Domain Methods (Keep logic inside Domain!)

func (u *User) UpdateName(firstName, lastName string) error {
	if strings.TrimSpace(firstName) == "" {
		return ErrInvalidFirstName
	}
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
}

// Validation Helpers
func looksLikeEmail(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}
