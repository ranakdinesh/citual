package app

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/ranakdinesh/citual/internal/platform/config"
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
	// setting up the database
	//dbPool := db.NewPool(ctx, cfg.DatabaseURL)
	routerSetup := func(r chi.Router) {

	}
	srv := httpserver.NewServer(httpserver.Options{
		Addr: cfg.HTTPAddr,
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
