package app

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/ranakdinesh/citual/internal/modules/identity"
	"github.com/ranakdinesh/citual/internal/platform/config"
	"github.com/ranakdinesh/citual/internal/platform/db"
	"github.com/ranakdinesh/citual/internal/platform/httpserver"
	"github.com/ranakdinesh/citual/internal/platform/logger"
)

type App struct {
	Config config.Config

	Logger *logger.Loggerx
	Server *httpserver.Server
}

func New(ctx context.Context) (*App, error) {
	// 1. Load Configuration
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		return nil, fmt.Errorf("app: failed to load config: %w", err)
	}

	// 2. Initialize Logging
	log := logger.NewWithOptions(logger.Options{
		Environment: cfg.AppEnv,
	})

	// 3. Database
	dbPool := db.NewPool(ctx, cfg.DatabaseURL)

	// 4. Modules
	identityModule := identity.New(dbPool, log)

	routerSetup := func(r chi.Router) {
		r.Mount("/", identityModule.Router)
		// r.Mount("/api/v1/identity", identityModule.Router) // Alternative
	}
	srv := httpserver.NewServer(httpserver.Options{
		Addr:       cfg.HTTPAddr,
		EnableCORS: true,
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type", "X-CSRF-Token",
		},
		AllowedOrigins: []string{"*"}, // Allow all for dev
	}, log, routerSetup)

	return &App{
		Config: cfg,
		Logger: log,
		Server: srv,
	}, nil
}
func (a *App) Start(ctx context.Context) error {
	return a.Server.Start(ctx)
}
