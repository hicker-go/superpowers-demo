// Package auth provides authentication and session management.
package auth

import (
	"context"

	"github.com/qinzj/superpowers-demo/internal/domain"
)

// SessionRepository defines persistence operations for HTTP sessions.
// Interface is defined in the consuming (service) layer per project architecture.
type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	// GetByTokenWithUser returns the session and its user if valid and not expired.
	GetByTokenWithUser(ctx context.Context, token string) (*domain.Session, *domain.User, error)
}
