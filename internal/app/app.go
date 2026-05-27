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
	engage "github.com/ranakdinesh/spur-engage"
	engagepermissions "github.com/ranakdinesh/spur-engage/pkg/permissions"
	identity "github.com/ranakdinesh/spur-identity"
	messaging "github.com/ranakdinesh/spur-messaging"
	"github.com/ranakdinesh/spur-messaging/pkg/authctx"
	msgpermissions "github.com/ranakdinesh/spur-messaging/pkg/permissions"
	storage "github.com/ranakdinesh/spur-storage"
	storagepermissions "github.com/ranakdinesh/spur-storage/pkg/permissions"
	template "github.com/ranakdinesh/spur-template"
	"github.com/rs/zerolog"
	// SPUR:IMPORTS:END
)

type App struct {
	Infra *infrastructure.Infra
	// SPUR:APP_VALUES
	Identity  *identity.Module
	Messaging *messaging.Module
	Storage   *storage.Module
	Template  *template.Module
	Engage    *engage.Module
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
		FrontendURL:       cfg.FrontendURL,
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
	engageLog := new(zerolog.Logger)
	*engageLog = infra.Log.Logger()
	engageTemplate := &engageTemplateAdapter{}
	engageBrandRepo := newIdentityEngageBrandRepository()
	engageModule, err := engage.New(ctx, engage.Options{
		DB:                infra.DB,
		Log:               engageLog,
		MessageDispatcher: engageTemplate,
		PublicLeadGuard:   newEngageRedisPublicLeadGuard(infra.Redis),
		CaptchaVerifier:   newEngageTurnstileVerifier(cfg.EngageTurnstileSecret),
		BrandRepository:   engageBrandRepo,
		Cfg: engage.Config{
			PublicBaseURL:     cfg.EngagePublicBaseURL,
			AIEnabled:         cfg.EngageAIEnabled,
			AutomationEnabled: cfg.EngageAutomationEnabled,
			CaptchaSiteKey:    cfg.EngageTurnstileSiteKey,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("engage: %w", err)
	}

	messagingCfg := messaging.Config{
		AppEnv:           cfg.AppEnv,
		EncryptionKey:    encKey,
		WebhookBaseURL:   cfg.MessagingWebhookBaseURL,
		DefaultRateLimit: rateLimit,

		WorkerCount:                    workerCount,
		EmailProvider:                  emailProvider,
		EmailAPIKey:                    emailAPIKey,
		EmailFromAddress:               emailFrom,
		EmailFromName:                  emailName,
		EmailTrackOpens:                cfg.MessagingEmailTrackOpens != "false",
		EmailTrackClicks:               cfg.MessagingEmailTrackClicks != "false",
		SMSProvider:                    smsProvider,
		SMSAPIKey:                      smsAPIKey,
		SMSSenderID:                    smsSender,
		WhatsAppWebhookVerifyToken:     cfg.WhatsAppWebhookVerifyToken,
		WhatsAppMetaAppID:              cfg.WhatsAppMetaAppID,
		WhatsAppMetaAppSecret:          cfg.WhatsAppMetaAppSecret,
		WhatsAppGraphAPIVersion:        cfg.WhatsAppGraphAPIVersion,
		WhatsAppEmbeddedSignupConfigID: cfg.WhatsAppEmbeddedSignupConfigID,
		WhatsAppAccessToken:            cfg.WhatsAppAccessToken,
		WhatsAppDefaultWABAID:          cfg.WhatsAppDefaultWABAID,
		WhatsAppDefaultPhoneNumberID:   cfg.WhatsAppDefaultPhoneNumberID,
		WhatsAppDefaultBusinessID:      cfg.WhatsAppDefaultBusinessID,
		WhatsAppDefaultDisplayPhone:    cfg.WhatsAppDefaultDisplayPhone,
	}
	messagingLog := new(zerolog.Logger)
	*messagingLog = infra.Log.Logger()
	messagingModule, err := messaging.New(ctx, messaging.Options{
		DB:          infra.DB,
		Log:         messagingLog,
		Cfg:         messagingCfg,
		Redis:       infra.Redis,
		EngageInbox: newMessagingEngageInboxAdapter(engageModule.Services.Inbox),
	})
	if err != nil {
		return nil, fmt.Errorf("messaging: %w", err)
	}

	storageLog := new(zerolog.Logger)
	*storageLog = infra.Log.Logger()
	storageModule, err := storage.New(ctx, storage.Options{
		DB:  infra.DB,
		Log: storageLog,
		Cfg: storage.Config{
			Provider:       cfg.StorageProvider,
			LocalRoot:      cfg.StorageLocalRoot,
			Bucket:         cfg.StorageBucket,
			DefaultQuota:   cfg.StorageDefaultQuotaGB * 1024 * 1024 * 1024,
			MaxUploadBytes: cfg.StorageMaxUploadMB * 1024 * 1024,
			S3Endpoint:     cfg.StorageS3Endpoint,
			S3Region:       cfg.StorageS3Region,
			S3AccessKey:    cfg.StorageS3AccessKey,
			S3SecretKey:    cfg.StorageS3SecretKey,
			S3UsePathStyle: cfg.StorageS3UsePathStyle,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("storage: %w", err)
	}

	templateModule, err := template.New(ctx, template.Options{
		Messaging: newTemplateMessagingAdapter(messagingModule.Services.MessagingGateway),
		Storage:   newTemplateStorageAdapter(storageModule.Services.Storage),
		Logger:    templateLogger{log: storageLog},
	})
	if err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}
	engageTemplate.services = templateModule.Services

	identityLog := new(zerolog.Logger)
	*identityLog = infra.Log.Logger()
	identityTemplate := newIdentityTemplateAdapter(templateModule.Services)
	identityModule, err := identity.New(ctx, identity.Options{
		DB:            infra.DB,
		Log:           identityLog,
		Cfg:           identityCfg,
		PrivateKey:    privateKey,
		Communication: identityTemplate,
		ProfileFiles:  identityTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("identity: %w", err)
	}
	engageBrandRepo.service = identityModule.Services.TenantBrandService
	if err := identityModule.Services.ModuleService.RegisterManifest(ctx, messagingIdentityManifest()); err != nil {
		return nil, fmt.Errorf("register messaging manifest: %w", err)
	}
	if err := identityModule.Services.ModuleService.RegisterManifest(ctx, storageIdentityManifest()); err != nil {
		return nil, fmt.Errorf("register storage manifest: %w", err)
	}
	if err := identityModule.Services.ModuleService.RegisterManifest(ctx, engageIdentityManifest()); err != nil {
		return nil, fmt.Errorf("register engage manifest: %w", err)
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
		r.Group(func(r chi.Router) {
			r.Use(identityModule.AuthMiddleware())
			r.Use(identityModule.TenantIsolation())
			storageModule.RegisterRoutes(r)
		})
		r.Group(func(r chi.Router) {
			r.Use(identityModule.AuthMiddleware())
			r.Use(identityModule.TenantIsolation())
			engageModule.RegisterRoutes(r)
		})
		registerPublicBrandAssetRoutes(r, engageModule, storageModule)
		engageModule.RegisterPublicRoutes(r)
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
		Storage:   storageModule,
		Template:  templateModule,
		Engage:    engageModule,
		// SPUR:APP_RETURN:END
	}, nil
}

func engageIdentityManifest() identity.Manifest {
	permissions := make([]identity.ManifestPermission, 0, len(engagepermissions.Catalog))
	for _, permission := range engagepermissions.Catalog {
		permissions = append(permissions, identity.ManifestPermission{
			Slug:        permission.Key,
			Description: permission.Description,
		})
	}
	roleTemplates := make([]identity.ManifestRoleTemplate, 0, len(engagepermissions.RoleTemplates))
	for _, template := range engagepermissions.RoleTemplates {
		roleTemplates = append(roleTemplates, identity.ManifestRoleTemplate{
			Code:        template.Code,
			Name:        template.Name,
			Description: template.Description,
			Permissions: append([]string(nil), template.Permissions...),
		})
	}
	return identity.Manifest{
		Name:          "Engage",
		Code:          engagepermissions.ModuleCode,
		Description:   "Customer engagement CRM, deals, tasks, quotes, inbox, meetings, automations, and AI-assisted business updates.",
		Permissions:   permissions,
		RoleTemplates: roleTemplates,
	}
}

func storageIdentityManifest() identity.Manifest {
	permissions := make([]identity.ManifestPermission, 0, len(storagepermissions.Catalog))
	for _, permission := range storagepermissions.Catalog {
		permissions = append(permissions, identity.ManifestPermission{
			Slug:        permission.Key,
			Description: permission.Description,
		})
	}
	roleTemplates := make([]identity.ManifestRoleTemplate, 0, len(storagepermissions.RoleTemplates))
	for _, template := range storagepermissions.RoleTemplates {
		roleTemplates = append(roleTemplates, identity.ManifestRoleTemplate{
			Code:        template.Code,
			Name:        template.Name,
			Description: template.Description,
			Permissions: append([]string(nil), template.Permissions...),
		})
	}
	return identity.Manifest{
		Name:          "Storage",
		Code:          storagepermissions.ModuleCode,
		Description:   "Tenant documents, user avatars, messaging media, imports, and secure object access.",
		Permissions:   permissions,
		RoleTemplates: roleTemplates,
	}
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

			key, err := identityModule.Services.APIKeyService.Authenticate(r.Context(), apiKeyValue, r.Header.Get("Origin"), requestIP(r))
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

func requestIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
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
