package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"y/internal/config"
	"y/internal/infrastructure"
	"y/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	identity "github.com/ranakdinesh/spur-identity"
	messaging "github.com/ranakdinesh/spur-messaging"
	"github.com/ranakdinesh/spur-messaging/pkg/authctx"
	msgpermissions "github.com/ranakdinesh/spur-messaging/pkg/permissions"
	"github.com/rs/zerolog"
	// SPUR:IMPORTS:END
)

type App struct {
	Infra *infrastructure.Infra
	// SPUR:APP_VALUES
	Identity  *identity.Module
	Messaging *messaging.Module
	// SPUR:APP_VALUES:END
}

func New(ctx context.Context) (*App, error) {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	log := logger.NewWithOptions(logger.Options{
		Environment: cfg.AppEnv,
	})

	infra, err := infrastructure.Bootstrap(ctx, &cfg, log)
	if err != nil {
		return nil, err
	}

	// SPUR:MODULES
	authClientID, err := uuid.Parse(cfg.AuthClientID)
	if err != nil {
		return nil, fmt.Errorf("AUTH_CLIENT_ID: %w", err)
	}
	privateKey, err := config.LoadPrivateKey(cfg.JWTPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("JWT private key: %w", err)
	}
	identityCfg := identity.Config{
		Issuer:            cfg.IdentityIssuer,
		GlobalSecret:      []byte(cfg.FositeGlobalSecret),
		JWTPrivateKeyPath: cfg.JWTPrivateKeyPath,
		AuthClientId:      authClientID,
		AuthClientSecret:  cfg.AuthClientSecret,
		CookieName:        "spur_sso",
		CookieSecure:      cfg.AppEnv == "production",
		BootstrapKey:      cfg.APIKeyValue,
	}

	encKey, err := hex.DecodeString(cfg.MessagingEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("MESSAGING_ENCRYPTION_KEY must be 64-char hex: %w", err)
	}
	if len(encKey) != 32 {
		return nil, fmt.Errorf("MESSAGING_ENCRYPTION_KEY must be 32 bytes (64 hex chars)")
	}
	rateLimit := 10
	if cfg.MessagingDefaultRateLimit != "" {
		rateLimit, _ = strconv.Atoi(cfg.MessagingDefaultRateLimit)
	}
	workerCount := 5
	if cfg.MessagingWorkerCount != "" {
		workerCount, _ = strconv.Atoi(cfg.MessagingWorkerCount)
	}
	emailProvider := cfg.MessagingEmailProvider
	if emailProvider == "" {
		emailProvider = "sendgrid"
	}
	var emailAPIKey string
	switch emailProvider {
	case "sendgrid":
		emailAPIKey = cfg.SendGridAPIKey
	case "mailgun":
		emailAPIKey = cfg.MailgunAPIKey
	case "postmark":
		emailAPIKey = cfg.PostmarkServerToken
	}
	smsProvider := cfg.MessagingSMSProvider
	if smsProvider == "" {
		smsProvider = "msg91"
	}
	var smsAPIKey string
	switch smsProvider {
	case "msg91":
		smsAPIKey = cfg.MSG91AuthKey
	case "twilio":
		smsAPIKey = cfg.TwilioAccountSID
	}
	emailFrom := cfg.MessagingEmailFromAddr
	if emailFrom == "" {
		emailFrom = "noreply@example.com"
	}
	emailName := cfg.MessagingEmailFromName
	if emailName == "" {
		emailName = "Spur"
	}
	smsSender := cfg.MessagingSMSSenderID
	if smsSender == "" {
		smsSender = "SPUR"
	}
	messagingCfg := messaging.Config{
		AppEnv:           cfg.AppEnv,
		EncryptionKey:    encKey,
		WebhookBaseURL:   cfg.MessagingWebhookBaseURL,
		DefaultRateLimit: rateLimit,

		WorkerCount:                workerCount,
		EmailProvider:              emailProvider,
		EmailAPIKey:                emailAPIKey,
		EmailFromAddress:           emailFrom,
		EmailFromName:              emailName,
		EmailTrackOpens:            cfg.MessagingEmailTrackOpens != "false",
		EmailTrackClicks:           cfg.MessagingEmailTrackClicks != "false",
		SMSProvider:                smsProvider,
		SMSAPIKey:                  smsAPIKey,
		SMSSenderID:                smsSender,
		WhatsAppWebhookVerifyToken: cfg.WhatsAppWebhookVerifyToken,
		WhatsAppMetaAppID:          cfg.WhatsAppMetaAppID,
	}
	messagingLog := new(zerolog.Logger)
	*messagingLog = infra.Log.Logger()
	messagingModule, err := messaging.New(ctx, messaging.Options{
		DB:    infra.DB,
		Log:   messagingLog,
		Cfg:   messagingCfg,
		Redis: infra.Redis,
	})
	if err != nil {
		return nil, fmt.Errorf("messaging: %w", err)
	}

	identityLog := new(zerolog.Logger)
	*identityLog = infra.Log.Logger()
	identityModule, err := identity.New(ctx, identity.Options{
		DB:            infra.DB,
		Log:           identityLog,
		Cfg:           identityCfg,
		PrivateKey:    privateKey,
		Communication: newIdentityMessagingAdapter(messagingModule.Services.MessagingGateway, messagingLog),
	})
	if err != nil {
		return nil, fmt.Errorf("identity: %w", err)
	}
	if err := identityModule.Services.ModuleService.RegisterManifest(ctx, messagingIdentityManifest()); err != nil {
		return nil, fmt.Errorf("register messaging manifest: %w", err)
	}
	// SPUR:MODULES:END

	infra.HTTP.Mount(func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})
		// SPUR:ROUTES
		identityModule.RegisterRoutes(r)
		r.Route("/messaging", func(r chi.Router) {
			r.Use(messagingAuthMiddleware(identityModule))
			r.Use(identityModule.TenantIsolation())
			messagingModule.RegisterRoutes(r)
		})
		r.Get("/messaging/webhook/whatsapp", messagingModule.WebhookHandler.Verify)
		r.Post("/messaging/webhook/whatsapp", messagingModule.WebhookHandler.Handle)
		r.Post("/messaging/webhook/email/sendgrid", messagingModule.WebhookHandler.HandleSendGrid)
		r.Post("/messaging/webhook/email/mailgun", messagingModule.WebhookHandler.HandleMailgun)
		r.Post("/messaging/webhook/email/postmark", messagingModule.WebhookHandler.HandlePostmark)
		r.Post("/messaging/webhook/sms", messagingModule.WebhookHandler.HandleSMS)
		r.Get("/messaging/unsubscribe/{token}", messagingModule.WebhookHandler.HandleUnsubscribeLink)
		// SPUR:ROUTES:END
	})

	return &App{
		Infra: infra,
		// SPUR:APP_RETURN
		Identity:  identityModule,
		Messaging: messagingModule,
		// SPUR:APP_RETURN:END
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	return a.Infra.HTTP.Start(ctx)
}

func messagingIdentityManifest() identity.Manifest {
	permissions := make([]identity.ManifestPermission, 0, len(msgpermissions.Catalog))
	for _, permission := range msgpermissions.Catalog {
		permissions = append(permissions, identity.ManifestPermission{
			Slug:        permission.Key,
			Description: permission.Description,
		})
	}

	roleTemplates := make([]identity.ManifestRoleTemplate, 0, len(msgpermissions.RoleTemplates))
	for _, template := range msgpermissions.RoleTemplates {
		roleTemplates = append(roleTemplates, identity.ManifestRoleTemplate{
			Code:        template.Code,
			Name:        template.Name,
			Description: template.Description,
			Permissions: append([]string(nil), template.Permissions...),
		})
	}

	return identity.Manifest{
		Name:          "Messaging",
		Code:          msgpermissions.ModuleCode,
		Description:   "Messaging channels, templates, contacts, campaigns, delivery reports, and analytics.",
		Permissions:   permissions,
		RoleTemplates: roleTemplates,
	}
}

func messagingAuthMiddleware(identityModule *identity.Module) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		jwtMiddleware := identityModule.AuthMiddleware()
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKeyValue := messagingAPIKey(r)
			if apiKeyValue == "" {
				jwtMiddleware(next).ServeHTTP(w, r)
				return
			}

			key, err := identityModule.Services.APIKeyService.Authenticate(r.Context(), apiKeyValue, r.Header.Get("Origin"))
			if err != nil {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}

			ctx := identity.ContextWithAPIKey(r.Context(), key)
			ctx = authctx.WithAPIKey(ctx, key.TenantID, key.Scopes)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func messagingAPIKey(r *http.Request) string {
	if key := strings.TrimSpace(r.Header.Get("X-API-Key")); key != "" {
		return key
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(auth, "ApiKey ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "ApiKey "))
	}
	if strings.HasPrefix(auth, "Bearer ") {
		value := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if strings.Count(value, ".") != 2 {
			return value
		}
	}
	return ""
}
