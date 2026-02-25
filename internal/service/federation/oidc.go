package federation

import (
	"context"

	"github.com/qinzj/superpowers-demo/internal/domain"
)

// UpstreamUserInfo holds claims from the upstream IdP's UserInfo endpoint.
type UpstreamUserInfo struct {
	Sub               string
	Email             string
	PreferredUsername string
}

// OIDCExchange exchanges an authorization code for tokens and fetches user info from upstream IdP.
type OIDCExchange interface {
	ExchangeAndUserInfo(ctx context.Context, connector *domain.IdPConnector, code, redirectURI string) (*UpstreamUserInfo, error)
}
