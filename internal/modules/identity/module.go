package identity

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/fosite_store"
	identityHttp "github.com/ranakdinesh/citual/internal/modules/identity/adapters/http"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/http/handlers"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/oauth2"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/postgres"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/services"
	"github.com/ranakdinesh/citual/internal/platform/logger"
)

type Module struct {
	Router chi.Router
}

func New(pool *pgxpool.Pool, log *logger.Loggerx) *Module {
	store := postgres.NewStore(pool)

	// 1. Fosite Store (Adapter)
	fositeStore := fosite_store.NewStore(store)

	// 2. OAuth2 Provider
	// Secret Key should come from config. For now using a hardcoded placeholder or env.
	// TODO: Move to config
	secretKey := "some-secret-key-at-least-32-chars-long"
	provider, err := oauth2.NewProvider(fositeStore, log.Logger().With().Str("component", "fosite").Logger(), secretKey)
	if err != nil {
		// Log fatal or panic since we can't start auth module without it
		log.Fatal(context.Background()).Err(err).Msg("Failed to initialize OAuth2 Provider")
	}

	// Services
	fositeSvc := services.NewFositeService(store, provider)
	regSvc := services.NewRegistrationService(store, log)
	authSvc := services.NewAuthService(store, store, log)

	// Handlers
	clientHandler := handlers.NewClientHandler(fositeSvc)
	regHandler := handlers.NewRegistrationHandler(regSvc)
	authHandler := handlers.NewAuthHandler(authSvc, fositeSvc)

	r := chi.NewRouter()
	identityHttp.RegisterRoutes(r, clientHandler, regHandler, authHandler)

	return &Module{
		Router: r,
	}
}
