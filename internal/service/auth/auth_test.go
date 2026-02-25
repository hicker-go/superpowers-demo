package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/qinzj/superpowers-demo/ent/enttest"
	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
	"github.com/qinzj/superpowers-demo/internal/storage"
)

func TestAuthService_ValidateCredentials(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	userRepo := storage.NewUserRepository(client)
	authSvc := NewAuthService(userRepo)

	ctx := context.Background()
	hash, err := password.Hash("secret123")
	require.NoError(t, err)

	u := &domain.User{
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	err = userRepo.Create(ctx, u)
	require.NoError(t, err)

	t.Run("valid_credentials_returns_user", func(t *testing.T) {
		got, err := authSvc.ValidateCredentials(ctx, "alice", "secret123")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "alice", got.Username)
		require.Equal(t, u.ID, got.ID)
	})

	t.Run("wrong_password_returns_ErrInvalidCredentials", func(t *testing.T) {
		got, err := authSvc.ValidateCredentials(ctx, "alice", "wrong")
		require.Nil(t, got)
		require.True(t, errors.Is(err, ErrInvalidCredentials))
	})

	t.Run("user_not_found_returns_ErrInvalidCredentials", func(t *testing.T) {
		got, err := authSvc.ValidateCredentials(ctx, "nobody", "any")
		require.Nil(t, got)
		require.True(t, errors.Is(err, ErrInvalidCredentials))
	})
}
