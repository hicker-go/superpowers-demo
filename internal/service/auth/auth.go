package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
	"github.com/qinzj/superpowers-demo/internal/service/user"
)

// ErrInvalidCredentials is returned when username or password is invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

const sessionTokenBytes = 32
const sessionDuration = 24 * time.Hour

// AuthService provides authentication and credential validation.
type AuthService struct {
	userRepo    user.UserRepository
	sessionRepo SessionRepository
}

// NewAuthService creates an AuthService with the given repositories.
func NewAuthService(userRepo user.UserRepository, sessionRepo SessionRepository) *AuthService {
	return &AuthService{userRepo: userRepo, sessionRepo: sessionRepo}
}

// ValidateCredentials checks username and password against stored user.
// Returns the user if valid, ErrInvalidCredentials for wrong credentials, or an error on failure.
func (s *AuthService) ValidateCredentials(ctx context.Context, username, pwd string) (*domain.User, error) {
	u, err := s.userRepo.ByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("validate credentials: %w", err)
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}
	if !password.Verify(pwd, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}
