package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/qinzj/superpowers-demo/ent"
	"github.com/qinzj/superpowers-demo/ent/user"
	"github.com/qinzj/superpowers-demo/internal/domain"
)

// UserRepository implements user.UserRepository using ent.
type UserRepository struct {
	client *ent.Client
}

// NewUserRepository creates a UserRepository backed by the given ent client.
func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

// Create persists the user and populates u.ID with the generated ID.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	entUser, err := r.client.User.Create().
		SetUsername(u.Username).
		SetEmail(u.Email).
		SetPasswordHash(u.PasswordHash).
		SetCreatedAt(u.CreatedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	u.ID = strconv.Itoa(entUser.ID)
	return nil
}

// ByUsername returns the user with the given username, or nil if not found.
func (r *UserRepository) ByUsername(ctx context.Context, username string) (*domain.User, error) {
	entUser, err := r.client.User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by username: %w", err)
	}
	return entUserToDomain(entUser), nil
}

func entUserToDomain(e *ent.User) *domain.User {
	return &domain.User{
		ID:           strconv.Itoa(e.ID),
		Username:     e.Username,
		Email:        e.Email,
		PasswordHash: e.PasswordHash,
		CreatedAt:    e.CreatedAt,
	}
}
