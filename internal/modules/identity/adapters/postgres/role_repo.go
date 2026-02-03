package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/domain"
)

func (s *Store) CreateRole(ctx context.Context, r *domain.Role) (*domain.Role, error) {
	arg := sqlc.CreateRoleParams{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
	}
	row, err := s.getQueries(ctx).CreateRole(ctx, arg)
	if err != nil {
		return nil, err
	}
	return mapRoleToDomain(row), nil
}

func (s *Store) GetRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Role, error) {
	row, err := s.getQueries(ctx).GetRole(ctx, sqlc.GetRoleParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return mapRoleToDomain(row), nil
}

func (s *Store) GetRoleByCode(ctx context.Context, code string, tenantID uuid.UUID) (*domain.Role, error) {
	row, err := s.getQueries(ctx).GetRoleByCode(ctx, sqlc.GetRoleByCodeParams{
		Code:     &code,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return mapRoleToDomain(row), nil
}

func (s *Store) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	rows, err := s.getQueries(ctx).ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Role, len(rows))
	for i, r := range rows {
		result[i] = mapRoleToDomain(r)
	}
	return result, nil
}

func (s *Store) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.getQueries(ctx).AssignRoleToUser(ctx, sqlc.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (s *Store) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.getQueries(ctx).RemoveRoleFromUser(ctx, sqlc.RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (s *Store) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	rows, err := s.getQueries(ctx).ListUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Role, len(rows))
	for i, r := range rows {
		result[i] = mapRoleToDomain(r)
	}
	return result, nil
}

func (s *Store) DeleteRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.getQueries(ctx).DeleteRole(ctx, sqlc.DeleteRoleParams{
		ID:       id,
		TenantID: tenantID,
	})
}

func mapRoleToDomain(row sqlc.Roles) *domain.Role {
	return &domain.Role{
		ID:          row.ID,
		TenantID:    row.TenantID,
		Name:        row.Name,
		Code:        row.Code,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
	}
}
