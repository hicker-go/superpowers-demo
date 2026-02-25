package user

import (
	"context"

	"github.com/qinzj/superpowers-demo/internal/domain"
)

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
