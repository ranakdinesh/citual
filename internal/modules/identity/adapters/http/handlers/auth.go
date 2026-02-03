package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ory/fosite"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/ports"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/services"
	"github.com/ranakdinesh/citual/internal/platform/httpserver"
)

type AuthHandler struct {
	authSvc   *services.AuthService
	fositeSvc *services.FositeService
}

func NewAuthHandler(authSvc *services.AuthService, fositeSvc *services.FositeService) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		fositeSvc: fositeSvc,
	}
}

// Login API - Authenticates user and sets SSO cookie
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req ports.LoginCmd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()

	session, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		// Differentiate errors if needed (401 vs 500)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set SSO Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "citual_sso",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set to true in prod
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Authorize Endpoint (OAuth2)
func (h *AuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Initialize Fosite Context
	ar, err := h.fositeSvc.NewAuthorizeRequest(ctx, r)
	if err != nil {
		h.fositeSvc.WriteAuthorizeError(ctx, w, ar, err)
		return
	}

	// 2. Check SSO Session
	cookie, err := r.Cookie("citual_sso")
	if err != nil || cookie.Value == "" {
		// Log the error for debugging
		fmt.Printf("Authorize: Missing Cookie. Err: %v\n", err)

		// No session -> Redirect to Frontend Login Page
		http.Redirect(w, r, "http://localhost:3000/login?return_to="+url.QueryEscape(r.RequestURI), http.StatusFound)
		return
	}

	fmt.Printf("Authorize: Received Cookie: %s\n", cookie.Value)

	decodedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		fmt.Printf("Authorize: Cookie Decode Error: %v\n", err)
		http.Redirect(w, r, "http://localhost:3000/login?return_to="+url.QueryEscape(r.RequestURI), http.StatusFound)
		return
	}

	session, err := h.authSvc.GetSession(ctx, decodedValue)
	if err != nil {
		// Log the error
		fmt.Printf("Authorize: Invalid Session. Err: %v\n", err)

		// Invalid session -> Redirect to Login
		http.Redirect(w, r, "http://localhost:3000/login?return_to="+url.QueryEscape(r.RequestURI), http.StatusFound)
		return
	}

	// 3. Approve Request
	// In a real world scenario, you might show a consent screen here if it's a 3rd party client.
	// For 1st party, we auto-approve.
	userData := &ports.SessionUserData{
		UserID:   session.UserID.String(),
		TenantID: session.TenantID.String(),
		// Add Roles if needed in token
	}

	// CRITICAL FIX: Inject TenantID into Context
	// Fosite Store implementation requires TenantID to be present in context to create the session.
	// Since this request comes from a redirect (no headers), we must manually inject it from the valid SSO session.

	tid := session.TenantID.String()
	fmt.Printf("Authorize: Injecting TenantID: %s\n", tid)
	ctx = httpserver.SetTenantID(ctx, tid)

	resp, err := h.fositeSvc.NewAuthorizeResponse(ctx, ar, userData)
	if err != nil {
		fmt.Printf("Authorize: NewAuthorizeResponse Error: %v\n", err)
		if fositeErr, ok := err.(*fosite.RFC6749Error); ok {
			fmt.Printf("Authorize: Fosite Debug: %v\n", fositeErr.DebugField)
			fmt.Printf("Authorize: Fosite Hint: %v\n", fositeErr.HintField)
		}
		h.fositeSvc.WriteAuthorizeError(ctx, w, ar, err)
		return
	}

	// 4. Write Response (Redirect with Code)
	h.fositeSvc.WriteAuthorizeResponse(ctx, w, ar, resp)
}

// Token Endpoint (OAuth2)
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Create Access Request
	ar, err := h.fositeSvc.NewAccessRequest(ctx, r)
	if err != nil {
		fmt.Printf("Token: NewAccessRequest Error: %v\n", err)
		if fositeErr, ok := err.(*fosite.RFC6749Error); ok {
			fmt.Printf("Token: Fosite Debug: %v\n", fositeErr.DebugField)
			fmt.Printf("Token: Fosite Hint: %v\n", fositeErr.HintField)
		}
		h.fositeSvc.WriteAccessError(ctx, w, ar, err)
		return
	}

	// PATCH: Helper to ensure TenantID is in context for the Store creation
	// Since Token Endpoint is backend-to-backend (or client-to-backend), the context doesn't have the User's TenantID yet.
	// But `ar` has the Session loaded from the Authorization Code.
	if sess, ok := ar.GetSession().(*services.CitualSession); ok {
		// Try Claims.Extra first (where we put it)
		if tid, ok := sess.Claims.Extra["tenant_id"].(string); ok && tid != "" {
			ctx = httpserver.SetTenantID(ctx, tid)
		} else {
			fmt.Printf("Token: Warning - TenantID missing in CitualSession Claims Extra: %v\n", sess.Claims.Extra)
		}
	} else {
		// This should not happen if fosite_service is correct
		fmt.Printf("Token: Warning - Session is not services.CitualSession, got type: %T\n", ar.GetSession())
	}

	// 2. Create Access Response
	resp, err := h.fositeSvc.NewAccessResponse(ctx, ar)
	if err != nil {
		fmt.Printf("Token: NewAccessResponse Error: %v\n", err)
		if fositeErr, ok := err.(*fosite.RFC6749Error); ok {
			fmt.Printf("Token: Fosite Response Debug: %v\n", fositeErr.DebugField)
			fmt.Printf("Token: Fosite Response Hint: %v\n", fositeErr.HintField)
			fmt.Printf("Token: Fosite Response Description: %v\n", fositeErr.DescriptionField)
			if fositeErr.Cause() != nil {
				fmt.Printf("Token: Fosite Response Cause: %v\n", fositeErr.Cause())
			}
		}
		h.fositeSvc.WriteAccessError(ctx, w, ar, err)
		return
	}

	// 3. Write Response (JSON Token)
	h.fositeSvc.WriteAccessResponse(ctx, w, ar, resp)
}
