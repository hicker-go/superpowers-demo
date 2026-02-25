package user

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/qinzj/superpowers-demo/ent/enttest"
	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
	"github.com/qinzj/superpowers-demo/internal/storage"
)

func TestUserService_Create(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	repo := storage.NewUserRepository(client)
	svc := NewUserService(repo)

	ctx := context.Background()
	u := &domain.User{
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		CreatedAt:    time.Now(),
	}

	err := svc.Create(ctx, u)
	require.NoError(t, err)
	require.NotEmpty(t, u.ID, "Create should populate ID")

	found, err := repo.ByUsername(ctx, "alice")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, u.ID, found.ID)
	require.Equal(t, "alice", found.Username)
	require.Equal(t, "alice@example.com", found.Email)
	require.Equal(t, "hashed", found.PasswordHash)
}

func TestUserService_Register(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	repo := storage.NewUserRepository(client)
	svc := NewUserService(repo)
	ctx := context.Background()

	u, err := svc.Register(ctx, "bob", "bob@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, u)
	require.NotEmpty(t, u.ID)
	require.Equal(t, "bob", u.Username)
	require.Equal(t, "bob@example.com", u.Email)
	require.NotEmpty(t, u.PasswordHash)
	require.NotEqual(t, "password123", u.PasswordHash)

	// Duplicate username
	_, err = svc.Register(ctx, "bob", "other@example.com", "otherpass1")
	require.ErrorIs(t, err, ErrUsernameTaken)

	// Weak password
	_, err = svc.Register(ctx, "carol", "carol@example.com", "short")
	require.ErrorIs(t, err, ErrWeakPassword)

	// Verify stored user can be looked up
	found, err := repo.ByUsername(ctx, "bob")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.True(t, password.Verify("password123", found.PasswordHash))
}
