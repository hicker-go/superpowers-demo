package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/qinzj/superpowers-demo/internal/domain"
)

func generateSessionToken() (string, error) {
	b := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession creates a new HTTP session for the given user and returns it.
func (s *AuthService) CreateSession(ctx context.Context, userID string) (*domain.Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	expiresAt := time.Now().Add(sessionDuration)
	sess := &domain.Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

// GetSession returns the user associated with the given session token if valid.
func (s *AuthService) GetSession(ctx context.Context, token string) (*domain.User, error) {
	_, u, err := s.sessionRepo.GetByTokenWithUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return u, nil
}
