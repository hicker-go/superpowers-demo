package federation

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/qinzj/superpowers-demo/ent/enttest"
	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/storage"
)

// fakeOIDCExchange returns predetermined userinfo for testing.
type fakeOIDCExchange struct {
	userInfo *UpstreamUserInfo
	err     error
}

func (f *fakeOIDCExchange) ExchangeAndUserInfo(_ context.Context, _ *domain.IdPConnector, _, _ string) (*UpstreamUserInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.userInfo, nil
}

func TestFederationService_LoginWithUpstream(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	userRepo := storage.NewUserRepository(client)
	connectorRepo := storage.NewIdPConnectorRepository(client)
	sessionRepo := storage.NewSessionRepository(client)
	authSvc := auth.NewAuthService(userRepo, sessionRepo)

	ctx := context.Background()

	// Create IdP connector in DB
	entConn, err := client.IdPConnector.Create().
		SetIssuer("https://idp.example.com").
		SetClientID("test-client").
		SetClientSecret("secret").
		Save(ctx)
	require.NoError(t, err)
	connectorID := fmt.Sprintf("%d", entConn.ID)

	fakeOIDC := &fakeOIDCExchange{
		userInfo: &UpstreamUserInfo{
			Sub:               "upstream-sub-123",
			Email:             "federated@example.com",
			PreferredUsername: "federateduser",
		},
	}

	svc := NewFederationService(connectorRepo, fakeOIDC, userRepo, authSvc)

	t.Run("creates_new_user_and_session_when_email_not_found", func(t *testing.T) {
		callbackURL := fmt.Sprintf("http://localhost/auth/callback/%s", connectorID)
		sess, err := svc.LoginWithUpstream(ctx, connectorID, "state-ok", "auth-code", callbackURL)
		require.NoError(t, err)
		require.NotNil(t, sess)
		require.NotEmpty(t, sess.Token)
		require.True(t, sess.ExpiresAt.After(time.Now()))

		// User should exist
		u, err := userRepo.ByEmail(ctx, "federated@example.com")
		require.NoError(t, err)
		require.NotNil(t, u)
		require.Equal(t, "federateduser", u.Username)
		require.Equal(t, "federated@example.com", u.Email)
	})

	t.Run("returns_ErrConnectorNotFound_when_connector_missing", func(t *testing.T) {
		sess, err := svc.LoginWithUpstream(ctx, "99999", "state", "code", "http://localhost/callback")
		require.Nil(t, sess)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrConnectorNotFound)
	})

	t.Run("links_existing_user_by_email_and_creates_session", func(t *testing.T) {
		// Pre-create user with same email
		hash := "placeholder-hash"
		existing := &domain.User{
			Username:     "existing",
			Email:        "existing@example.com",
			PasswordHash: hash,
			CreatedAt:    time.Now(),
		}
		err := userRepo.Create(ctx, existing)
		require.NoError(t, err)

		fakeOIDC.userInfo = &UpstreamUserInfo{
			Sub:               "other-sub",
			Email:             "existing@example.com",
			PreferredUsername: "existing",
		}

		callbackURL := fmt.Sprintf("http://localhost/auth/callback/%s", connectorID)
		sess, err := svc.LoginWithUpstream(ctx, connectorID, "state-ok", "auth-code", callbackURL)
		require.NoError(t, err)
		require.NotNil(t, sess)
		require.Equal(t, existing.ID, sess.UserID)
	})
}
