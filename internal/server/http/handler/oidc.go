// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"

	"github.com/qinzj/superpowers-demo/internal/service/oidc"
)

// OIDCHandler handles OIDC/OAuth2 endpoints by delegating to Fosite.
type OIDCHandler struct {
	Provider fosite.OAuth2Provider
	Issuer   string
}

// NewOIDCHandler creates an OIDC handler with the given provider and issuer.
func NewOIDCHandler(provider fosite.OAuth2Provider, issuer string) *OIDCHandler {
	return &OIDCHandler{Provider: provider, Issuer: issuer}
}

// WellKnown serves GET /.well-known/openid-configuration (OIDC discovery).
func (h *OIDCHandler) WellKnown(c *gin.Context) {
	issuer := h.Issuer
	if issuer == "" {
		issuer = oidc.DefaultIssuerFromRequest(c.Request)
	}
	doc := oidc.DiscoveryDocument(issuer)
	c.JSON(http.StatusOK, doc)
}

// Authorize handles GET /authorize. Validates the request and redirects to login
// when no session exists, or writes the authorize response (302) when session exists.
func (h *OIDCHandler) Authorize(c *gin.Context) {
	ctx := c.Request.Context()
	ar, err := h.Provider.NewAuthorizeRequest(ctx, c.Request)
	if err != nil {
		h.Provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
		return
	}

	if !hasSession(c) {
		redirectToLogin(c, ar)
		return
	}

	session := sessionFromContext(c)
	response, err := h.Provider.NewAuthorizeResponse(ctx, ar, session)
	if err != nil {
		h.Provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
		return
	}
	h.Provider.WriteAuthorizeResponse(ctx, c.Writer, ar, response)
}

// Token handles POST /token. Returns JSON with access_token, token_type, etc.
func (h *OIDCHandler) Token(c *gin.Context) {
	ctx := c.Request.Context()
	session := new(fosite.DefaultSession)
	accessRequest, err := h.Provider.NewAccessRequest(ctx, c.Request, session)
	if err != nil {
		h.Provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
		return
	}

	response, err := h.Provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		h.Provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
		return
	}
	h.Provider.WriteAccessResponse(ctx, c.Writer, accessRequest, response)
}

// UserInfo handles GET /userinfo. Validates Bearer token and returns user claims as JSON.
func (h *OIDCHandler) UserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	token := fosite.AccessTokenFromRequest(c.Request)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return
	}

	session := new(fosite.DefaultSession)
	_, ar, err := h.Provider.IntrospectToken(ctx, token, fosite.AccessToken, session, "openid")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	claims := userInfoClaims(ar)
	c.JSON(http.StatusOK, claims)
}

func hasSession(c *gin.Context) bool {
	_, err := c.Cookie("sso_session")
	return err == nil
}

func sessionFromContext(c *gin.Context) *fosite.DefaultSession {
	// TODO(Task 11): Load session from cookie/storage.
	return &fosite.DefaultSession{
		Username: "placeholder",
		Subject:  "placeholder",
	}
}

func redirectToLogin(c *gin.Context, ar fosite.AuthorizeRequester) {
	loginURL, _ := url.Parse("/login")
	q := loginURL.Query()
	q.Set("client_id", ar.GetClient().GetID())
	if ru := ar.GetRedirectURI(); ru != nil {
		q.Set("redirect_uri", ru.String())
	}
	q.Set("response_type", strings.Join(ar.GetResponseTypes(), " "))
	q.Set("scope", strings.Join(ar.GetRequestedScopes(), " "))
	q.Set("state", ar.GetState())
	loginURL.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, loginURL.String())
}

func userInfoClaims(ar fosite.AccessRequester) map[string]interface{} {
	claims := make(map[string]interface{})
	sess := ar.GetSession()
	if sess == nil {
		return claims
	}
	if sub := sess.GetSubject(); sub != "" {
		claims["sub"] = sub
	}
	if username := sess.GetUsername(); username != "" {
		claims["preferred_username"] = username
	}
	if ext, ok := sess.(interface{ GetExtraClaims() map[string]interface{} }); ok {
		for k, v := range ext.GetExtraClaims() {
			claims[k] = v
		}
	}
	return claims
}
