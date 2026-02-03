package oauth2

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/rs/zerolog"
)

// Provider wraps the Fosite library interface
type Provider interface {
	fosite.OAuth2Provider
}

// NewProvider creates a Fosite instance with Refresh Token support and correct V0.49+ signatures
func NewProvider(storage fosite.Storage, logger zerolog.Logger, secretKey string) (fosite.OAuth2Provider, error) {

	// 1. Configuration
	config := &fosite.Config{
		AccessTokenLifespan:  time.Hour * 1,
		RefreshTokenLifespan: time.Hour * 24 * 30, // 30 Days
		GlobalSecret:         []byte(secretKey),
		// Critical: Allow "offline" scope for Refresh Tokens
		RefreshTokenScopes: []string{"offline", "offline_access"},
		// Use Argon2 for Client Secrets
		ClientSecretsHasher: NewArgon2Hasher(),
		IDTokenIssuer:       "http://localhost:8090", // Match backend URL
	}

	// 2. Load Private Key
	privateKey, err := loadOrCreatePrivateKey("private_key.pem")
	if err != nil {
		return nil, err
	}

	// 3. Define the KeyGetter
	keyGetter := func(ctx context.Context) (interface{}, error) {
		return privateKey, nil
	}

	// 4. Create Strategies
	// HMAC is used as an internal helper for JWT strategy construction.
	hmacStrategy := compose.NewOAuth2HMACStrategy(config)
	jwtStrategy := compose.NewOAuth2JWTStrategy(keyGetter, hmacStrategy, config)
	oidcStrategy := compose.NewOpenIDConnectStrategy(keyGetter, config)

	// 5. The Composition
	return compose.Compose(
		config,
		storage,
		&compose.CommonStrategy{
			CoreStrategy:               jwtStrategy,
			OpenIDConnectTokenStrategy: oidcStrategy,
		},
		// --- ENABLED HANDLERS ---

		// 1. OAuth2 Authorization Code
		compose.OAuth2AuthorizeExplicitFactory,

		// 2. OAuth2 Refresh Token
		compose.OAuth2RefreshTokenGrantFactory,

		// 3. OpenID Connect
		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectRefreshFactory,

		// 4. PKCE
		compose.OAuth2PKCEFactory,

		// 5. Client Credentials
		compose.OAuth2ClientCredentialsGrantFactory,
	), nil
}

// Helper: Load or Generate RSA Key
func loadOrCreatePrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		block, _ := pem.Decode(b)
		if block != nil {
			return x509.ParsePKCS1PrivateKey(block.Bytes)
		}
	}

	// Generate new if missing (Dev only - in Prod use vault)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return key, nil
}
