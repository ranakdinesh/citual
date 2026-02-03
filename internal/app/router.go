package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates a new Chi router
func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Mount module sub-routers here

	return r
}
