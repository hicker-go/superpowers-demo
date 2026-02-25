// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/service/federation"
)

// FederationHandler handles federation init (redirect to upstream IdP).
type FederationHandler struct {
	Federation *federation.FederationService
	Issuer     string
}

// NewFederationHandler creates a FederationHandler with the given service and issuer.
func NewFederationHandler(f *federation.FederationService, issuer string) *FederationHandler {
	return &FederationHandler{Federation: f, Issuer: issuer}
}

// Init handles GET /auth/federation/:connector_id. Encodes OAuth params into state and
// redirects the user to the upstream IdP authorize URL.
func (h *FederationHandler) Init(c *gin.Context) {
	connectorID := c.Param("connector_id")
	if connectorID == "" {
		c.Redirect(http.StatusFound, "/login?error=invalid_connector")
		return
	}

	params := FederationParams{
		ClientID:     c.Query("client_id"),
		RedirectURI:  c.Query("redirect_uri"),
		ResponseType: c.Query("response_type"),
		Scope:        c.Query("scope"),
		State:        c.Query("state"),
	}
	stateJSON, err := json.Marshal(params)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=invalid_params")
		return
	}
	stateB64 := base64.URLEncoding.EncodeToString(stateJSON)

	ctx := c.Request.Context()
	authURL, err := h.Federation.AuthCodeURL(ctx, connectorID, h.Issuer, stateB64)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=connector_not_found")
		return
	}

	c.Redirect(http.StatusFound, authURL)
}
