package user

import (
	"context"

	"github.com/qinzj/superpowers-demo/internal/domain"
)

// UserRepository defines persistence operations for users.
// Interface is defined in the consuming (service) layer per project architecture.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ByUsername(ctx context.Context, username string) (*domain.User, error)
}
