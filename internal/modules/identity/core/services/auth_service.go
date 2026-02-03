package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ranakdinesh/citual/internal/modules/identity/core/ports"
	"github.com/ranakdinesh/citual/internal/platform/logger"
	"github.com/ranakdinesh/citual/internal/platform/security"
)

type AuthService struct {
	userRepo    ports.UserRepo
	sessionRepo ports.SessionRepo
	log         *logger.Loggerx
}

func NewAuthService(userRepo ports.UserRepo, sessionRepo ports.SessionRepo, log *logger.Loggerx) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		log:         log,
	}
}

func (s *AuthService) Login(ctx context.Context, cmd ports.LoginCmd) (*ports.Session, error) {
	s.log.Info(ctx).Str("email", cmd.Email).Msg("Attempting login")

	// 1. Find User
	user, err := s.userRepo.GetUserByEmail(ctx, cmd.Email)
	if err != nil {
		s.log.Warn(ctx).Str("email", cmd.Email).Err(err).Msg("Login failed: user not found")
		return nil, errors.New("invalid credentials")
	}

	// 2. Refresh User data (to ensure we have password hash if not retrieved by Email query - usually it is)
	// Assuming GetUserByEmail returns PasswordHash.

	// 3. Verify Password (Argon2)
	match, err := security.VerifyPassword(cmd.Password, user.PasswordHash)
	if err != nil {
		s.log.Error(ctx).Err(err).Msg("Login failed: password check error")
		return nil, errors.New("internal server error")
	}
	if !match {
		s.log.Warn(ctx).Str("email", cmd.Email).Msg("Login failed: invalid password")
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}

	// 4. Create Session
	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	session := &ports.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 1 Day Session
	}

	err = s.sessionRepo.CreateUserSession(ctx, session)
	if err != nil {
		s.log.Error(ctx).Err(err).Msg("Failed to create session")
		return nil, errors.New("internal server error")
	}

	s.log.Info(ctx).Str("user_id", user.ID.String()).Msg("Login successful")
	return session, nil
}

func (s *AuthService) GetSession(ctx context.Context, token string) (*ports.Session, error) {
	session, err := s.sessionRepo.GetUserSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session lazily
		_ = s.sessionRepo.DeleteUserSession(ctx, token)
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteUserSession(ctx, token)
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
