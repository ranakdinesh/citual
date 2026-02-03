package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TenantKind string

const (
	TenantKindOps      TenantKind = "ops"
	TenantKindCustomer TenantKind = "customer"
)

var (
	ErrInvalidTenantName = errors.New("invalid tenant name")
)

type Tenant struct {
	ID               uuid.UUID
	Name             string
	Kind             TenantKind
	TrialEndsAt      *time.Time
	SubscriptionPlan *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewTenant(name string, kind TenantKind) (*Tenant, error) {
	if name == "" {
		return nil, ErrInvalidTenantName
	}

	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Kind:      kind,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
