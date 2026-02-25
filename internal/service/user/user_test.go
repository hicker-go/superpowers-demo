package user

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/qinzj/superpowers-demo/ent/enttest"
	"github.com/qinzj/superpowers-demo/internal/domain"
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
