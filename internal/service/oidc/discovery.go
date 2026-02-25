// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package oidc

import (
	"net/http"
	"strings"
)

// DiscoveryDocument returns the OIDC discovery metadata for the given issuer.
func DiscoveryDocument(issuer string) map[string]interface{} {
	// Strip trailing slash for consistent URLs.
	base := strings.TrimSuffix(issuer, "/")
	return map[string]interface{}{
		"issuer":                                base,
		"authorization_endpoint":                base + "/authorize",
		"token_endpoint":                        base + "/token",
		"userinfo_endpoint":                     base + "/userinfo",
		"jwks_uri":                             base + "/jwks.json",
		"scopes_supported":                     []string{"openid", "profile", "email", "offline_access"},
		"response_types_supported":             []string{"code", "token", "id_token", "code token", "code id_token", "id_token token", "code id_token token"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "implicit"},
		"subject_types_supported":              []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                     []string{"sub", "iss", "aud", "exp", "iat", "email", "name", "preferred_username"},
	}
}

// DefaultIssuerFromRequest derives an issuer URL from the HTTP request (scheme + host).
func DefaultIssuerFromRequest(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8888"
	}
	return scheme + "://" + host
}
