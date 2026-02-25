// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/token/jwt"
)

// OIDCConfig holds OIDC provider configuration.
type OIDCConfig struct {
	Issuer             string
	AccessTokenLifespan time.Duration
	RefreshTokenLifespan time.Duration
	IDTokenLifespan     time.Duration
	GlobalSecret       []byte
	PrivateKey         *rsa.PrivateKey
}

// DefaultOIDCConfig returns config with sensible defaults.
func DefaultOIDCConfig(issuer string) (*OIDCConfig, error) {
	// GlobalSecret must be exactly 32 bytes for HMAC signing.
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &OIDCConfig{
		Issuer:               issuer,
		AccessTokenLifespan:  30 * time.Minute,
		RefreshTokenLifespan: 24 * time.Hour,
		IDTokenLifespan:      1 * time.Hour,
		GlobalSecret:         secret,
		PrivateKey:           key,
	}, nil
}

// NewFositeConfig builds fosite.Config for the OIDC provider.
func (c *OIDCConfig) NewFositeConfig() *fosite.Config {
	return &fosite.Config{
		AccessTokenLifespan:       c.AccessTokenLifespan,
		RefreshTokenLifespan:      c.RefreshTokenLifespan,
		IDTokenLifespan:           c.IDTokenLifespan,
		GlobalSecret:              c.GlobalSecret,
		IDTokenIssuer:             c.Issuer,
		AccessTokenIssuer:         c.Issuer,
		ScopeStrategy:             fosite.HierarchicScopeStrategy,
		AudienceMatchingStrategy:  fosite.DefaultAudienceMatchingStrategy,
	}
}

// NewOAuth2Provider creates a Fosite OAuth2/OIDC provider with all standard handlers.
func NewOAuth2Provider(cfg *OIDCConfig, storage fosite.Storage) fosite.OAuth2Provider {
	config := cfg.NewFositeConfig()
	keyGetter := func(context.Context) (interface{}, error) {
		return cfg.PrivateKey, nil
	}
	return compose.ComposeAllEnabled(
		config,
		storage,
		keyGetter,
	)
}

// GetSigner returns a JWT signer for the config.
func (c *OIDCConfig) GetSigner() *jwt.DefaultSigner {
	return &jwt.DefaultSigner{
		GetPrivateKey: func(context.Context) (interface{}, error) {
			return c.PrivateKey, nil
		},
	}
}
