// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/service/federation"
)

// FederationParams holds OAuth params encoded in the state for federation flow.
type FederationParams struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	ResponseType string `json:"response_type"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
}

// CallbackHandler handles the upstream IdP OAuth callback.
type CallbackHandler struct {
	Federation *federation.FederationService
	Issuer     string
}

// NewCallbackHandler creates a CallbackHandler with the given federation service and issuer.
func NewCallbackHandler(f *federation.FederationService, issuer string) *CallbackHandler {
	return &CallbackHandler{Federation: f, Issuer: issuer}
}

// GetCallback handles GET /auth/callback/:connector_id. Parses code and state, exchanges with
// upstream IdP, creates session, and redirects to /authorize to continue the OIDC flow.
func (h *CallbackHandler) GetCallback(c *gin.Context) {
	connectorID := c.Param("connector_id")
	code := c.Query("code")
	stateB64 := c.Query("state")

	if connectorID == "" || code == "" || stateB64 == "" {
		c.Redirect(http.StatusFound, "/login?error=invalid_callback_params")
		return
	}

	var params FederationParams
	stateJSON, err := base64.URLEncoding.DecodeString(stateB64)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=invalid_state")
		return
	}
	if err := json.Unmarshal(stateJSON, &params); err != nil {
		c.Redirect(http.StatusFound, "/login?error=invalid_state")
		return
	}

	redirectURI := buildCallbackRedirectURI(h.Issuer, connectorID)
	ctx := c.Request.Context()
	sess, err := h.Federation.LoginWithUpstream(ctx, connectorID, stateB64, code, redirectURI)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=federation_failed")
		return
	}

	c.SetCookie(sessionCookieName, sess.Token, sessionCookieMaxAge, "/", "", false, true)

	if params.ClientID == "" && params.RedirectURI == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	authURL := buildAuthorizeURL(LoginParams{
		ClientID:     params.ClientID,
		RedirectURI:  params.RedirectURI,
		ResponseType: params.ResponseType,
		Scope:        params.Scope,
		State:        params.State,
	})
	c.Redirect(http.StatusFound, authURL)
}

func buildCallbackRedirectURI(issuer string, connectorID string) string {
	return strings.TrimSuffix(issuer, "/") + "/auth/callback/" + connectorID
}
