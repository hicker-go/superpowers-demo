package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
)

// ErrUsernameTaken is returned when the username is already in use.
var ErrUsernameTaken = errors.New("username already taken")

// ErrWeakPassword is returned when the password does not meet strength requirements.
var ErrWeakPassword = errors.New("password does not meet strength requirements")

const minPasswordLen = 8

// UserService provides user business operations.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a UserService with the given repository.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create persists a new user. The user's ID is populated after creation.
func (s *UserService) Create(ctx context.Context, u *domain.User) error {
	return s.repo.Create(ctx, u)
}

// Delete removes the user and all their sessions. Idempotent: returns nil if user already deleted.
func (s *UserService) Delete(ctx context.Context, userID string) error {
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// Register creates a new user after validating username uniqueness and password strength.
func (s *UserService) Register(ctx context.Context, username, email, pwd string) (*domain.User, error) {
	existing, err := s.repo.ByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}
	if len(pwd) < minPasswordLen {
		return nil, ErrWeakPassword
	}
	hash, err := password.Hash(pwd)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	if err := s.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}
