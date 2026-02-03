package domain

import (
	"time"

	"github.com/google/uuid"
)

type ModuleStatus string

const (
	ModuleStatusActive   ModuleStatus = "active"
	ModuleStatusDisabled ModuleStatus = "disabled"
)

// Represents a feature enabled for a specific tenant
type TenantModule struct {
	TenantID  uuid.UUID
	ModuleKey string // e.g. "crm"
	Status    ModuleStatus
	EnabledAt time.Time
}
