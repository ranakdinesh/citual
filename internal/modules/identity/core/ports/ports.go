package ports

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/postgres/sqlc"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/domain"
)

// Legacy Fosite Service (Keep as is for now or refactor later if time permits)
// For now, FositeService uses sqlc types because it was implemented first.
// Ideally it should also be domain-based. But let's focus on the new Registration flow first.
// BUT, mixing them is confusing.
// Let's refactor FositeService ports too if possible, but maybe minimal touch for now.
// Actually, `sqlc.FositeClients` is effectively a domain object for that sub-domain.
// Let's leave FositeService as is for this step to reduce risk, unless it breaks import cycles.

type ClientRepo interface {
	CreateClient(ctx context.Context, arg sqlc.CreateClientParams) (sqlc.FositeClients, error)
	GetClient(ctx context.Context, id string) (sqlc.FositeClients, error)
	GetActiveClient(ctx context.Context, id string) (sqlc.FositeClients, error)
	ListClients(ctx context.Context, tenantID uuid.UUID) ([]sqlc.FositeClients, error)
	ListPublicClients(ctx context.Context) ([]sqlc.FositeClients, error)
	UpdateClientSecret(ctx context.Context, arg sqlc.UpdateClientSecretParams) error
	ToggleClientStatus(ctx context.Context, arg sqlc.ToggleClientStatusParams) error
	UpdateClientConfig(ctx context.Context, arg sqlc.UpdateClientConfigParams) error
	DeleteClient(ctx context.Context, id string) error
}

type FositeSessionRepo interface {
	CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) error
	GetSession(ctx context.Context, arg sqlc.GetSessionParams) (sqlc.FositeSessions, error)
	DeleteSessionByType(ctx context.Context, arg sqlc.DeleteSessionByTypeParams) error
	RevokeSessionByRequestId(ctx context.Context, requestID string) error
	RevokeSessionByRequestIdAndType(ctx context.Context, arg sqlc.RevokeSessionByRequestIdAndTypeParams) error
}

type FositeService interface {
	CreateClient(ctx context.Context, cmd CreateClientCmd) (*sqlc.FositeClients, error)
	GetClient(ctx context.Context, id string) (*sqlc.FositeClients, error)
	ListClients(ctx context.Context, tenantID uuid.UUID) ([]*sqlc.FositeClients, error)
	UpdateClient(ctx context.Context, id string, cmd UpdateClientCmd) error
	DeleteClient(ctx context.Context, id string) error
	ListPublicClients(ctx context.Context) ([]*sqlc.FositeClients, error)

	// OAuth2 Handlers
	NewAuthorizeRequest(ctx context.Context, r *http.Request) (fosite.AuthorizeRequester, error)
	NewAuthorizeResponse(ctx context.Context, ar fosite.AuthorizeRequester, session *SessionUserData) (fosite.AuthorizeResponder, error)
	WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, ar fosite.AuthorizeRequester, resp fosite.AuthorizeResponder)
	WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, ar fosite.AuthorizeRequester, err error)

	NewAccessRequest(ctx context.Context, r *http.Request) (fosite.AccessRequester, error)
	NewAccessResponse(ctx context.Context, ar fosite.AccessRequester) (fosite.AccessResponder, error)
	WriteAccessResponse(ctx context.Context, rw http.ResponseWriter, ar fosite.AccessRequester, resp fosite.AccessResponder)
	WriteAccessError(ctx context.Context, rw http.ResponseWriter, ar fosite.AccessRequester, err error)
}

type CreateClientCmd struct {
	TenantID      uuid.UUID
	ClientSecret  *string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	Scopes        []string
	Audience      []string
	Public        bool
}

type UpdateClientCmd struct {
	ClientSecret  *string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	Scopes        []string
	Audience      []string
	Active        *bool
}

// New Repositories (Registration Flow) - USING DOMAIN TYPES

type TenantRepo interface {
	CreateTenant(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	ListTenants(ctx context.Context) ([]*domain.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *domain.Tenant) error
}

type UserRepo interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

type RoleRepo interface {
	CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error)
	GetRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Role, error)
	GetRoleByCode(ctx context.Context, code string, tenantID uuid.UUID) (*domain.Role, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error)

	// These might remain specific assignments, but ideally domain concepts:
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// Registration Service

type RegistrationService interface {
	RegisterTenant(ctx context.Context, cmd RegisterTenantCmd) (*RegisteredTenantResult, error)
	CreateUser(ctx context.Context, cmd CreateUserCmd) (*domain.User, error)
}

type RegisterTenantCmd struct {
	FirstName   string
	LastName    string
	CompanyName string
	Email       string
	Mobile      string
	Password    string
}

type RegisteredTenantResult struct {
	Tenant *domain.Tenant
	User   *domain.User
}

type CreateUserCmd struct {
	TenantID     uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	Mobile       string
	Password     string
	Roles        []string
	IsSuperAdmin bool
}

type LoginCmd struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	TenantID  uuid.UUID
}

type SessionUserData struct {
	UserID   string
	TenantID string
}

type AuthService interface {
	Login(ctx context.Context, cmd LoginCmd) (*Session, error)
	GetSession(ctx context.Context, token string) (*Session, error)
	Logout(ctx context.Context, token string) error
}

type SessionRepo interface {
	CreateUserSession(ctx context.Context, s *Session) error
	GetUserSessionByToken(ctx context.Context, token string) (*Session, error)
	DeleteUserSession(ctx context.Context, token string) error
	DeleteUserSessionsByUserID(ctx context.Context, userID uuid.UUID) error
}
