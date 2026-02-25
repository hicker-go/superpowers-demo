// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
)

// OIDCRouteConfig holds OIDC handler configuration for route registration.
type OIDCRouteConfig struct {
	Provider fosite.OAuth2Provider
	Issuer   string
}

// NewEngine creates a new Gin engine. Routes are registered by the router package.
func NewEngine() *gin.Engine {
	return gin.Default()
}

// RegisterOIDCRoutes adds OIDC endpoints to the given engine.
func RegisterOIDCRoutes(e *gin.Engine, cfg *OIDCRouteConfig) {
	if cfg == nil || cfg.Provider == nil {
		return
	}
	h := NewOIDCHandler(cfg.Provider, cfg.Issuer)
	e.GET("/.well-known/openid-configuration", h.WellKnown)
	e.GET("/authorize", h.Authorize)
	e.POST("/token", h.Token)
	e.GET("/userinfo", h.UserInfo)
}
