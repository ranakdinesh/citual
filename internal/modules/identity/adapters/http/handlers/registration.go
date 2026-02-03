package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ranakdinesh/citual/internal/modules/identity/core/ports"
)

type RegistrationHandler struct {
	Service ports.RegistrationService
}

func NewRegistrationHandler(service ports.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{Service: service}
}

// RegisterTenant - Public Self-Registration
func (h *RegistrationHandler) RegisterTenant(w http.ResponseWriter, r *http.Request) {
	var cmd ports.RegisterTenantCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if cmd.Email == "" || cmd.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.Service.RegisterTenant(r.Context(), cmd)
	if err != nil {
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// The result contains *domain.Tenant and *domain.User, which should serialize to JSON correctly
	json.NewEncoder(w).Encode(result)
}

// RegisterUser - Admin creates user (Ops or Tenant Admin)
func (h *RegistrationHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var cmd ports.CreateUserCmd
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.Service.CreateUser(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
