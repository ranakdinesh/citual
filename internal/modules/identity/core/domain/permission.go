package domain

import "github.com/google/uuid"

type Permission struct {
	ID          uuid.UUID
	Module      string // e.g., "crm", "identity"
	Action      string // e.g., "write", "read"
	Slug        string // e.g., "crm:contact:write"
	Description string
}

// Reconstructs the slug from parts
func (p Permission) String() string {
	return p.Slug
}
