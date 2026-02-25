package federation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/oidc_client"
	"github.com/qinzj/superpowers-demo/internal/infra/password"
	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/user"
)

// ErrConnectorNotFound is returned when the IdP connector does not exist.
var ErrConnectorNotFound = errors.New("connector not found")

// FederationService handles upstream IdP login and identity linking.
type FederationService struct {
	connectorRepo IdPConnectorRepository
	oidcExchange  OIDCExchange
	userRepo      user.UserRepository
	authSvc       *auth.AuthService
}

// NewFederationService creates a FederationService with the given dependencies.
func NewFederationService(
	connectorRepo IdPConnectorRepository,
	oidcExchange OIDCExchange,
	userRepo user.UserRepository,
	authSvc *auth.AuthService,
) *FederationService {
	return &FederationService{
		connectorRepo: connectorRepo,
		oidcExchange: oidcExchange,
		userRepo:     userRepo,
		authSvc:      authSvc,
	}
}

// LoginWithUpstream exchanges the auth code for tokens, fetches userinfo from upstream IdP,
// maps/links identity to a local User, and creates a Session.
func (s *FederationService) LoginWithUpstream(
	ctx context.Context,
	connectorID, state, code, redirectURI string,
) (*domain.Session, error) {
	_ = state // state validation can be done at handler layer

	connector, err := s.connectorRepo.GetByID(ctx, connectorID)
	if err != nil {
		return nil, err
	}
	if connector == nil {
		return nil, ErrConnectorNotFound
	}

	userInfo, err := s.oidcExchange.ExchangeAndUserInfo(ctx, connector, code, redirectURI)
	if err != nil {
		return nil, err
	}

	// Resolve or create local user by email or upstream subject
	u, err := s.resolveOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	return s.authSvc.CreateSession(ctx, u.ID)
}

// ListConnectors returns all configured IdP connectors.
func (s *FederationService) ListConnectors(ctx context.Context) ([]*domain.IdPConnector, error) {
	return s.connectorRepo.List(ctx)
}

// AuthCodeURL returns the upstream IdP authorize URL for the given connector.
func (s *FederationService) AuthCodeURL(ctx context.Context, connectorID, issuer, state string) (string, error) {
	conn, err := s.connectorRepo.GetByID(ctx, connectorID)
	if err != nil {
		return "", fmt.Errorf("get connector: %w", err)
	}
	if conn == nil {
		return "", ErrConnectorNotFound
	}
	redirectURL := strings.TrimSuffix(issuer, "/") + "/auth/callback/" + connectorID
	client, err := oidc_client.NewClient(ctx, conn, redirectURL)
	if err != nil {
		return "", fmt.Errorf("create oidc client: %w", err)
	}
	return client.AuthCodeURL(state), nil
}

func (s *FederationService) resolveOrCreateUser(ctx context.Context, info *UpstreamUserInfo) (*domain.User, error) {
	// Try by email first
	if info.Email != "" {
		u, err := s.userRepo.ByEmail(ctx, info.Email)
		if err != nil {
			return nil, err
		}
		if u != nil {
			return u, nil
		}
	}

	// Create new user
	username := info.PreferredUsername
	if username == "" {
		username = info.Email
	}
	if username == "" {
		username = info.Sub
	}

	hash, err := federatedUserPasswordHash()
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Username:     username,
		Email:        info.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func federatedUserPasswordHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate federated placeholder: %w", err)
	}
	return password.Hash(hex.EncodeToString(b))
}
