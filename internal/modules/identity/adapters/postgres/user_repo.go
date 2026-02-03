package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/domain"
)

func (s *Store) CreateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	arg := sqlc.CreateUserParams{
		ID:           u.ID,
		TenantID:     u.TenantID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		Mobile:       nil,
		PasswordHash: u.PasswordHash,
		IsSuperAdmin: u.IsSuperAdmin,
		IsActive:     u.IsActive,
	}
	if u.Mobile != "" {
		arg.Mobile = &u.Mobile
	}

	row, err := s.getQueries(ctx).CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	return mapUserToDomain(row), nil
}

func (s *Store) GetUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.User, error) {
	row, err := s.getQueries(ctx).GetUser(ctx, sqlc.GetUserParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return mapUserToDomain(row), nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := s.getQueries(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapUserToDomain(row), nil
}

func (s *Store) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error) {
	rows, err := s.getQueries(ctx).ListUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.User, len(rows))
	for i, r := range rows {
		result[i] = mapUserToDomain(r)
	}
	return result, nil
}

func (s *Store) UpdateUser(ctx context.Context, u *domain.User) error {
	arg := sqlc.UpdateUserParams{
		ID:        u.ID,
		TenantID:  u.TenantID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
	if u.Mobile != "" {
		arg.Mobile = &u.Mobile
	}

	return s.getQueries(ctx).UpdateUser(ctx, arg)
}

func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.getQueries(ctx).DeleteUser(ctx, sqlc.DeleteUserParams{
		ID:       id,
		TenantID: tenantID,
	})
}

func mapUserToDomain(row sqlc.Users) *domain.User {
	u := &domain.User{
		ID:           row.ID,
		TenantID:     row.TenantID,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		IsSuperAdmin: row.IsSuperAdmin,
		AuthzVersion: int(row.AuthzVersion),
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.Mobile != nil {
		u.Mobile = *row.Mobile
	}
	if row.EmailVerifiedAt.Valid {
		val := row.EmailVerifiedAt.Time
		u.EmailVerifiedAt = &val
	}
	if row.MobileVerifiedAt.Valid {
		val := row.MobileVerifiedAt.Time
		u.MobileVerifiedAt = &val
	}
	return u
}
