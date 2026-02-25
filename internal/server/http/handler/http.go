// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"

	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/federation"
	"github.com/qinzj/superpowers-demo/internal/service/user"
)

// OIDCRouteConfig holds OIDC handler configuration for route registration.
type OIDCRouteConfig struct {
	Provider fosite.OAuth2Provider
	Issuer   string
	Auth     *auth.AuthService
}

// LoginRouteConfig holds login handler configuration.
type LoginRouteConfig struct {
	Auth        *auth.AuthService
	Federation  FederationRouteConfig
}

// FederationRouteConfig holds federation handler configuration.
type FederationRouteConfig struct {
	Service *federation.FederationService
	Issuer  string
}

// RegisterRouteConfig holds register handler configuration.
type RegisterRouteConfig struct {
	UserService *user.UserService
}

// NewEngine creates a new Gin engine with HTML templates. Routes are registered by the router package.
func NewEngine() *gin.Engine {
	e := gin.Default()
	e.LoadHTMLGlob(filepath.Join("internal", "server", "http", "templates", "*.html"))
	return e
}

// RegisterOIDCRoutes adds OIDC endpoints to the given engine.
func RegisterOIDCRoutes(e *gin.Engine, cfg *OIDCRouteConfig) {
	if cfg == nil || cfg.Provider == nil {
		return
	}
	h := NewOIDCHandler(cfg.Provider, cfg.Issuer, cfg.Auth)
	e.GET("/.well-known/openid-configuration", h.WellKnown)
	e.GET("/authorize", h.Authorize)
	e.POST("/token", h.Token)
	e.GET("/userinfo", h.UserInfo)
}

// RegisterLoginRoutes adds login endpoints to the given engine.
func RegisterLoginRoutes(e *gin.Engine, cfg *LoginRouteConfig) {
	if cfg == nil || cfg.Auth == nil {
		return
	}
	h := NewLoginHandler(cfg.Auth, cfg.Federation)
	e.GET("/login", h.GetLogin)
	e.POST("/login", h.PostLogin)
}

// RegisterFederationRoutes adds federation init and callback endpoints.
func RegisterFederationRoutes(e *gin.Engine, cfg *FederationRouteConfig) {
	if cfg == nil || cfg.Service == nil {
		return
	}
	fedH := NewFederationHandler(cfg.Service, cfg.Issuer)
	cbH := NewCallbackHandler(cfg.Service, cfg.Issuer)
	e.GET("/auth/federation/:connector_id", fedH.Init)
	e.GET("/auth/callback/:connector_id", cbH.GetCallback)
}

// RegisterRegisterRoutes adds registration endpoints to the given engine.
func RegisterRegisterRoutes(e *gin.Engine, cfg *RegisterRouteConfig) {
	if cfg == nil || cfg.UserService == nil {
		return
	}
	h := NewRegisterHandler(cfg.UserService)
	e.GET("/register", h.RegisterGet)
	e.POST("/register", h.RegisterPost)
}
