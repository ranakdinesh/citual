package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"y/internal/config"
	"y/internal/infrastructure"
	"y/internal/logger"
	"github.com/google/uuid"
	identity "github.com/ranakdinesh/spur-identity"
	"encoding/hex"
	"strconv"
	messaging "github.com/ranakdinesh/spur-messaging"
	// SPUR:IMPORTS:END
)

type App struct {
	Infra *infrastructure.Infra
	// SPUR:APP_VALUES
	Identity *identity.Module
	Messaging *messaging.Module
	// SPUR:APP_VALUES:END

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
if err != nil { return nil, fmt.Errorf("AUTH_CLIENT_ID: %w", err) }
privateKey, err := config.LoadPrivateKey(cfg.JWTPrivateKeyPath)
if err != nil { return nil, fmt.Errorf("JWT private key: %w", err) }
identityCfg := identity.Config{Issuer: cfg.IdentityIssuer, GlobalSecret: []byte(cfg.FositeGlobalSecret), JWTPrivateKeyPath: cfg.JWTPrivateKeyPath, AuthClientId: authClientID, AuthClientSecret: cfg.AuthClientSecret, CookieName: "spur_sso", CookieSecure: cfg.AppEnv == "production", BootstrapKey: cfg.APIKeyValue}
identityLog := infra.Log.Logger()
identityModule, err := identity.New(ctx, identity.Options{DB: infra.DB, Log: &identityLog, Cfg: identityCfg, PrivateKey: privateKey})
if err != nil { return nil, fmt.Errorf("identity: %w", err) }
}
	encKey, err := hex.DecodeString(cfg.MessagingEncryptionKey)
if err != nil { return nil, fmt.Errorf("MESSAGING_ENCRYPTION_KEY must be 64-char hex: %w", err) }
if len(encKey) != 32 { return nil, fmt.Errorf("MESSAGING_ENCRYPTION_KEY must be 32 bytes (64 hex chars)") }
rateLimit := 10
if cfg.MessagingDefaultRateLimit != "" { rateLimit, _ = strconv.Atoi(cfg.MessagingDefaultRateLimit) }
workerCount := 5
if cfg.MessagingWorkerCount != "" { workerCount, _ = strconv.Atoi(cfg.MessagingWorkerCount) }
emailProvider := cfg.MessagingEmailProvider
if emailProvider == "" { emailProvider = "sendgrid" }
var emailAPIKey string
switch emailProvider {
case "sendgrid": emailAPIKey = cfg.SendGridAPIKey
case "mailgun": emailAPIKey = cfg.MailgunAPIKey
case "postmark": emailAPIKey = cfg.PostmarkServerToken
}
smsProvider := cfg.MessagingSMSProvider
if smsProvider == "" { smsProvider = "msg91" }
var smsAPIKey string
switch smsProvider {
case "msg91": smsAPIKey = cfg.MSG91AuthKey
case "twilio": smsAPIKey = cfg.TwilioAccountSID
}
emailFrom := cfg.MessagingEmailFromAddr
if emailFrom == "" { emailFrom = "noreply@example.com" }
emailName := cfg.MessagingEmailFromName
if emailName == "" { emailName = "Spur" }
smsSender := cfg.MessagingSMSSenderID
if smsSender == "" { smsSender = "SPUR" }
messagingCfg := messaging.Config{EncryptionKey: encKey, WebhookBaseURL: cfg.MessagingWebhookBaseURL, DefaultRateLimit: rateLimit, RedisURL: cfg.MessagingRedisURL, WorkerCount: workerCount, EmailProvider: emailProvider, EmailAPIKey: emailAPIKey, EmailFromAddress: emailFrom, EmailFromName: emailName, EmailTrackOpens: cfg.MessagingEmailTrackOpens != "false", EmailTrackClicks: cfg.MessagingEmailTrackClicks != "false", SMSProvider: smsProvider, SMSAPIKey: smsAPIKey, SMSSenderID: smsSender, WhatsAppWebhookVerifyToken: cfg.WhatsAppWebhookVerifyToken, WhatsAppMetaAppID: cfg.WhatsAppMetaAppID}
messagingLog := infra.Log.Logger()
messagingModule, err := messaging.New(ctx, messaging.Options{DB: infra.DB, Log: &messagingLog, Cfg: messagingCfg})
if err != nil { return nil, fmt.Errorf("messaging: %%w", err) }
	// SPUR:MODULES:END

	infra.HTTP.Mount(func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})
		// SPUR:ROUTES
	identityModule.RegisterRoutes(r)
	r.Route("/messaging", func(r chi.Router) {
	r.Use(identityModule.AuthGuard.Middleware)
	r.Use(identityModule.AuthGuard.TenantIsolation)
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
	Identity: identityModule,
	Messaging: messagingModule,
		// SPUR:APP_RETURN:END
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	return a.Infra.HTTP.Start(ctx)
}
