package federation

import (
	"context"
	"fmt"

	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/infra/oidc_client"
)

// OIDCClientAdapter adapts infra oidc_client to the OIDCExchange interface.
type OIDCClientAdapter struct{}

// NewOIDCClientAdapter returns an adapter that uses the real OIDC client.
func NewOIDCClientAdapter() *OIDCClientAdapter {
	return &OIDCClientAdapter{}
}

// ExchangeAndUserInfo implements OIDCExchange using the infra oidc_client.
func (a *OIDCClientAdapter) ExchangeAndUserInfo(
	ctx context.Context,
	connector *domain.IdPConnector,
	code, redirectURI string,
) (*UpstreamUserInfo, error) {
	client, err := oidc_client.NewClient(ctx, connector, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("create oidc client: %w", err)
	}
	token, err := client.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %w", err)
	}
	info, err := client.UserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("userinfo: %w", err)
	}
	return &UpstreamUserInfo{
		Sub:               info.Sub,
		Email:             info.Email,
		PreferredUsername: info.PreferredUsername,
	}, nil
}
