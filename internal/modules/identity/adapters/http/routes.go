package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/ranakdinesh/citual/internal/modules/identity/adapters/http/handlers"
)

func RegisterRoutes(r chi.Router, clientHandler *handlers.ClientHandler, regHandler *handlers.RegistrationHandler, authHandler *handlers.AuthHandler) {
	// Public routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", regHandler.RegisterTenant)
		r.Post("/login", authHandler.Login)
	})

	r.Route("/oauth2", func(r chi.Router) {
		r.Get("/auth", authHandler.Authorize)
		r.Post("/token", authHandler.Token)
	})

	// Protected / Admin routes
	// Ideally under /admin or with middleware
	r.Post("/users", regHandler.RegisterUser)

	r.Route("/clients", func(r chi.Router) {
		r.Post("/", clientHandler.CreateClient)
		r.Get("/", clientHandler.ListClients)
		r.Get("/public", clientHandler.ListPublicClients)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", clientHandler.GetClient)
			r.Put("/", clientHandler.UpdateClient)
			r.Delete("/", clientHandler.DeleteClient)
		})
	})
}
