// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

// Package oidc_client provides OIDC/RP client for upstream identity providers.
package oidc_client

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/qinzj/superpowers-demo/internal/domain"
	"golang.org/x/oauth2"
)

// Client exchanges auth codes and fetches userinfo from upstream IdPs.
type Client struct {
	provider   *oidc.Provider
	oauth2Conf oauth2.Config
}

// NewClient creates an OIDC client for the given connector configuration.
func NewClient(ctx context.Context, conn *domain.IdPConnector, redirectURL string) (*Client, error) {
	provider, err := oidc.NewProvider(ctx, conn.Issuer)
	if err != nil {
		return nil, fmt.Errorf("create oidc provider: %w", err)
	}
	conf := oauth2.Config{
		ClientID:     conn.ClientID,
		ClientSecret: conn.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return &Client{provider: provider, oauth2Conf: conf}, nil
}

// AuthCodeURL returns the URL to redirect the user to for authorization.
func (c *Client) AuthCodeURL(state string) string {
	return c.oauth2Conf.AuthCodeURL(state)
}

// Exchange exchanges the authorization code for tokens.
func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.oauth2Conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	return token, nil
}

// UserInfo fetches user claims from id_token or the IdP's UserInfo endpoint.
func (c *Client) UserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	if rawIDToken, ok := token.Extra("id_token").(string); ok && rawIDToken != "" {
		verifier := c.provider.Verifier(&oidc.Config{ClientID: c.oauth2Conf.ClientID})
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return nil, fmt.Errorf("verify id token: %w", err)
		}
		var claims struct {
			Sub               string `json:"sub"`
			Email             string `json:"email"`
			PreferredUsername string `json:"preferred_username"`
			Name              string `json:"name"`
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, fmt.Errorf("parse claims: %w", err)
		}
		return &UserInfo{
			Sub:               claims.Sub,
			Email:             claims.Email,
			PreferredUsername: claims.PreferredUsername,
			Name:              claims.Name,
		}, nil
	}
	// Fallback to UserInfo endpoint when id_token is not present.
	oi, err := c.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("userinfo endpoint: %w", err)
	}
	var extra struct {
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	_ = oi.Claims(&extra)
	return &UserInfo{
		Sub:               oi.Subject,
		Email:             oi.Email,
		PreferredUsername: extra.PreferredUsername,
		Name:              extra.Name,
	}, nil
}

// UserInfo holds claims from upstream IdP userinfo.
type UserInfo struct {
	Sub               string
	Email             string
	PreferredUsername string
	Name              string
}
