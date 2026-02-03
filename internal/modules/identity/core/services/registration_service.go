package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/postgres"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/domain"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/ports"
	"github.com/ranakdinesh/citual/internal/platform/logger"
	"github.com/ranakdinesh/citual/internal/platform/security"
	//"golang.org/x/crypto/bcrypt"
)

type RegistrationService struct {
	store *postgres.Store
	log   *logger.Loggerx
}

func NewRegistrationService(store *postgres.Store, log *logger.Loggerx) *RegistrationService {
	return &RegistrationService{
		store: store,
		log:   log,
	}
}

// RegisterTenant performs transactional creation of Tenant + User + Role
func (s *RegistrationService) RegisterTenant(ctx context.Context, cmd ports.RegisterTenantCmd) (*ports.RegisteredTenantResult, error) {
	s.log.Info(ctx).Str("email", cmd.Email).Msg("Starting tenant registration")

	tenantName := cmd.CompanyName
	if tenantName == "" {
		tenantName = fmt.Sprintf("%s %s", cmd.FirstName, cmd.LastName)
	}

	// Password Hashing
	passwordHash, err := security.HashPassword(cmd.Password)
	if err != nil {
		s.log.Error(ctx).Err(err).Msg("failed to hash password")
		return nil, errors.New("internal server error")
	}

	var result ports.RegisteredTenantResult

	// Transactional Execution
	err = s.store.RunInTx(ctx, func(ctx context.Context) error {
		// 1. Create Tenant
		trialEnd := time.Now().Add(7 * 24 * time.Hour)
		subPlan := "trial"
		tenantModel := &domain.Tenant{
			ID:               uuid.New(),
			Name:             tenantName,
			Kind:             domain.TenantKindCustomer,
			TrialEndsAt:      &trialEnd,
			SubscriptionPlan: &subPlan,
		}

		tenant, err := s.store.CreateTenant(ctx, tenantModel)
		if err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}
		result.Tenant = tenant

		// 2. Create User
		// Use Domain Factory if possible, or manual struct
		userModel, err := domain.NewUser(tenant.ID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Mobile)
		if err != nil {
			return err
		}
		userModel.PasswordHash = passwordHash
		// IsSuperAdmin default false

		user, err := s.store.CreateUser(ctx, userModel)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		result.User = user

		// 3. Ensure TENANT_ADMIN role exists
		roleCode := "TENANT_ADMIN"
		role, err := s.store.GetRoleByCode(ctx, roleCode, tenant.ID)
		if err != nil {
			// Assume missing, create it.
			roleModel := &domain.Role{
				ID:          uuid.New(),
				TenantID:    tenant.ID,
				Name:        "Tenant Admin",
				Code:        &roleCode,
				Description: &roleCode,
			}
			role, err = s.store.CreateRole(ctx, roleModel)
			if err != nil {
				return fmt.Errorf("failed to create admin role: %w", err)
			}
		}

		// 4. Assign Role
		err = s.store.AssignRoleToUser(ctx, user.ID, role.ID)
		if err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		return nil
	})

	if err != nil {
		s.log.Error(ctx).Err(err).Msg("tenant registration failed")
		return nil, err
	}

	s.log.Info(ctx).Str("tenant_id", result.Tenant.ID.String()).Str("user_id", result.User.ID.String()).Msg("Tenant registration successful")
	return &result, nil
}

func (s *RegistrationService) CreateUser(ctx context.Context, cmd ports.CreateUserCmd) (*domain.User, error) {
	s.log.Info(ctx).Str("email", cmd.Email).Str("tenant_id", cmd.TenantID.String()).Msg("Creating user")

	passwordHash, err := security.HashPassword(cmd.Password)
	if err != nil {
		return nil, err
	}

	var createdUser *domain.User

	err = s.store.RunInTx(ctx, func(ctx context.Context) error {
		userModel, err := domain.NewUser(cmd.TenantID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Mobile)
		if err != nil {
			return err
		}
		userModel.PasswordHash = passwordHash
		userModel.IsSuperAdmin = cmd.IsSuperAdmin

		user, err := s.store.CreateUser(ctx, userModel)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		createdUser = user

		// Assign Roles
		for _, code := range cmd.Roles {
			role, err := s.store.GetRoleByCode(ctx, code, cmd.TenantID)
			if err != nil {
				return fmt.Errorf("role not found: %s", code)
			}
			if err := s.store.AssignRoleToUser(ctx, user.ID, role.ID); err != nil {
				return fmt.Errorf("failed to assign role %s: %w", code, err)
			}
		}

		return nil
	})

	if err != nil {
		s.log.Error(ctx).Err(err).Msg("create user failed")
		return nil, err
	}

	return createdUser, nil
}
