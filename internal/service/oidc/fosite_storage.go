// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package oidc

import (
	"context"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v3"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/handler/pkce"
	"github.com/ory/fosite/handler/rfc7523"

	"github.com/qinzj/superpowers-demo/ent"
	"github.com/qinzj/superpowers-demo/ent/oauth2client"
)

// FositeStorage implements fosite.Storage using ent for clients and in-memory for sessions.
// OAuth2Client entities are loaded from the database; authorize codes, access tokens, refresh
// tokens, PKCE, and OIDC sessions are stored in memory (suitable for development).
type FositeStorage struct {
	client *ent.Client

	// In-memory session stores (keyed by signature/code).
	authorizeCodes   map[string]storeAuthorizeCode
	accessTokens     map[string]fosite.Requester
	refreshTokens    map[string]storeRefreshToken
	oidcSessions     map[string]fosite.Requester
	pkceSessions     map[string]fosite.Requester
	accessTokenIDs   map[string]string // requestID -> signature
	refreshTokenIDs  map[string]string // requestID -> signature
	blacklistedJTIs  map[string]time.Time

	mu sync.RWMutex
}

type storeAuthorizeCode struct {
	active bool
	fosite.Requester
}

type storeRefreshToken struct {
	active               bool
	accessTokenSignature string
	fosite.Requester
}

// NewFositeStorage creates a FositeStorage backed by the given ent client.
func NewFositeStorage(client *ent.Client) *FositeStorage {
	return &FositeStorage{
		client:          client,
		authorizeCodes:  make(map[string]storeAuthorizeCode),
		accessTokens:    make(map[string]fosite.Requester),
		refreshTokens:   make(map[string]storeRefreshToken),
		oidcSessions:    make(map[string]fosite.Requester),
		pkceSessions:   make(map[string]fosite.Requester),
		accessTokenIDs:  make(map[string]string),
		refreshTokenIDs: make(map[string]string),
		blacklistedJTIs: make(map[string]time.Time),
	}
}

// Ensure FositeStorage implements required interfaces.
var (
	_ fosite.Storage                              = (*FositeStorage)(nil)
	_ oauth2.CoreStorage                          = (*FositeStorage)(nil)
	_ oauth2.TokenRevocationStorage               = (*FositeStorage)(nil)
	_ oauth2.ResourceOwnerPasswordCredentialsGrantStorage = (*FositeStorage)(nil)
	_ openid.OpenIDConnectRequestStorage           = (*FositeStorage)(nil)
	_ pkce.PKCERequestStorage                     = (*FositeStorage)(nil)
	_ rfc7523.RFC7523KeyStorage                   = (*FositeStorage)(nil)
)

// GetClient loads the OAuth2 client by ID from the database.
func (s *FositeStorage) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	c, err := s.client.OAuth2Client.Query().
		Where(oauth2client.ClientIDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}
	return entClientToFosite(c), nil
}

// ClientAssertionJWTValid returns nil if the JTI is not known (valid to use).
func (s *FositeStorage) ClientAssertionJWTValid(_ context.Context, jti string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if exp, exists := s.blacklistedJTIs[jti]; exists && exp.After(time.Now()) {
		return fosite.ErrJTIKnown
	}
	return nil
}

// SetClientAssertionJWT marks a JTI as used.
func (s *FositeStorage) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for j, e := range s.blacklistedJTIs {
		if e.Before(time.Now()) {
			delete(s.blacklistedJTIs, j)
		}
	}
	if _, exists := s.blacklistedJTIs[jti]; exists {
		return fosite.ErrJTIKnown
	}
	s.blacklistedJTIs[jti] = exp
	return nil
}

// CreateAuthorizeCodeSession stores the authorization code session.
func (s *FositeStorage) CreateAuthorizeCodeSession(_ context.Context, code string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authorizeCodes[code] = storeAuthorizeCode{active: true, Requester: req}
	return nil
}

// GetAuthorizeCodeSession retrieves the authorization code session.
func (s *FositeStorage) GetAuthorizeCodeSession(_ context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rel, ok := s.authorizeCodes[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	if !rel.active {
		return rel.Requester, fosite.ErrInvalidatedAuthorizeCode
	}
	return rel.Requester, nil
}

// InvalidateAuthorizeCodeSession invalidates an authorization code after use.
func (s *FositeStorage) InvalidateAuthorizeCodeSession(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rel, ok := s.authorizeCodes[code]
	if !ok {
		return fosite.ErrNotFound
	}
	rel.active = false
	s.authorizeCodes[code] = rel
	return nil
}

// CreateAccessTokenSession stores the access token session.
func (s *FositeStorage) CreateAccessTokenSession(_ context.Context, sig string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessTokens[sig] = req
	s.accessTokenIDs[req.GetID()] = sig
	return nil
}

// GetAccessTokenSession retrieves the access token session.
func (s *FositeStorage) GetAccessTokenSession(_ context.Context, sig string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.accessTokens[sig]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return req, nil
}

// DeleteAccessTokenSession deletes the access token session.
func (s *FositeStorage) DeleteAccessTokenSession(_ context.Context, sig string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req, ok := s.accessTokens[sig]; ok {
		delete(s.accessTokenIDs, req.GetID())
	}
	delete(s.accessTokens, sig)
	return nil
}

// CreateRefreshTokenSession stores the refresh token session.
func (s *FositeStorage) CreateRefreshTokenSession(_ context.Context, sig, accessSig string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[sig] = storeRefreshToken{
		active:               true,
		Requester:             req,
		accessTokenSignature: accessSig,
	}
	s.refreshTokenIDs[req.GetID()] = sig
	return nil
}

// GetRefreshTokenSession retrieves the refresh token session.
func (s *FositeStorage) GetRefreshTokenSession(_ context.Context, sig string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rel, ok := s.refreshTokens[sig]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	if !rel.active {
		return rel.Requester, fosite.ErrInactiveToken
	}
	return rel.Requester, nil
}

// DeleteRefreshTokenSession deletes the refresh token session.
func (s *FositeStorage) DeleteRefreshTokenSession(_ context.Context, sig string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rel, ok := s.refreshTokens[sig]; ok {
		delete(s.refreshTokenIDs, rel.GetID())
	}
	delete(s.refreshTokens, sig)
	return nil
}

// RotateRefreshToken revokes the old refresh token and access token before rotation.
func (s *FositeStorage) RotateRefreshToken(ctx context.Context, requestID, _ string) error {
	if err := s.RevokeRefreshToken(ctx, requestID); err != nil {
		return err
	}
	return s.RevokeAccessToken(ctx, requestID)
}

// RevokeAccessToken revokes an access token by request ID.
func (s *FositeStorage) RevokeAccessToken(ctx context.Context, requestID string) error {
	s.mu.RLock()
	sig, exists := s.accessTokenIDs[requestID]
	s.mu.RUnlock()
	if !exists {
		return nil
	}
	return s.DeleteAccessTokenSession(ctx, sig)
}

// RevokeRefreshToken revokes a refresh token by request ID.
func (s *FositeStorage) RevokeRefreshToken(_ context.Context, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sig, exists := s.refreshTokenIDs[requestID]; exists {
		if rel, ok := s.refreshTokens[sig]; ok {
			rel.active = false
			s.refreshTokens[sig] = rel
		}
	}
	return nil
}

// Authenticate implements ResourceOwnerPasswordCredentialsGrantStorage.
// Returns ErrNotFound; password grant is not supported in this minimal implementation.
func (s *FositeStorage) Authenticate(_ context.Context, _, _ string) (string, error) {
	return "", fosite.ErrNotFound
}

// CreateOpenIDConnectSession stores the OIDC session for an authorize code.
func (s *FositeStorage) CreateOpenIDConnectSession(_ context.Context, code string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.oidcSessions[code] = req
	return nil
}

// GetOpenIDConnectSession retrieves the OIDC session.
func (s *FositeStorage) GetOpenIDConnectSession(_ context.Context, code string, _ fosite.Requester) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.oidcSessions[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return req, nil
}

// DeleteOpenIDConnectSession deletes the OIDC session.
func (s *FositeStorage) DeleteOpenIDConnectSession(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.oidcSessions, code)
	return nil
}

// CreatePKCERequestSession stores the PKCE session.
func (s *FositeStorage) CreatePKCERequestSession(_ context.Context, code string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pkceSessions[code] = req
	return nil
}

// GetPKCERequestSession retrieves the PKCE session.
func (s *FositeStorage) GetPKCERequestSession(_ context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.pkceSessions[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return req, nil
}

// DeletePKCERequestSession deletes the PKCE session.
func (s *FositeStorage) DeletePKCERequestSession(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pkceSessions, code)
	return nil
}

// GetPublicKey implements RFC7523KeyStorage; returns ErrNotFound (JWT assertion grant not supported).
func (s *FositeStorage) GetPublicKey(_ context.Context, _, _, _ string) (*jose.JSONWebKey, error) {
	return nil, fosite.ErrNotFound
}

// GetPublicKeys implements RFC7523KeyStorage; returns ErrNotFound.
func (s *FositeStorage) GetPublicKeys(_ context.Context, _, _ string) (*jose.JSONWebKeySet, error) {
	return nil, fosite.ErrNotFound
}

// GetPublicKeyScopes implements RFC7523KeyStorage; returns ErrNotFound.
func (s *FositeStorage) GetPublicKeyScopes(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, fosite.ErrNotFound
}

// IsJWTUsed delegates to ClientAssertionJWTValid.
func (s *FositeStorage) IsJWTUsed(ctx context.Context, jti string) (bool, error) {
	err := s.ClientAssertionJWTValid(ctx, jti)
	if err != nil {
		return true, nil
	}
	return false, nil
}

// MarkJWTUsedForTime delegates to SetClientAssertionJWT.
func (s *FositeStorage) MarkJWTUsedForTime(ctx context.Context, jti string, exp time.Time) error {
	return s.SetClientAssertionJWT(ctx, jti, exp)
}

// entClientToFosite converts an ent OAuth2Client to fosite.DefaultClient.
// The client_secret in the DB should be stored as a bcrypt hash.
func entClientToFosite(c *ent.OAuth2Client) *fosite.DefaultClient {
	redirectURIs := c.RedirectUris
	if redirectURIs == nil {
		redirectURIs = []string{}
	}
	return &fosite.DefaultClient{
		ID:           c.ClientID,
		Secret:       []byte(c.ClientSecret),
		RedirectURIs:  redirectURIs,
		GrantTypes:   []string{"authorization_code", "refresh_token", "implicit"},
		ResponseTypes: []string{"code", "token", "id_token", "id_token token", "code id_token", "code token", "code id_token token"},
		Scopes:       []string{"openid", "profile", "email", "offline"},
	}
}
