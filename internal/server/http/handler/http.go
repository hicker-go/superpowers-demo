// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"go.uber.org/zap"

	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/federation"
	"github.com/qinzj/superpowers-demo/internal/service/user"
	"github.com/qinzj/superpowers-demo/pkg/log"
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

// AccountRouteConfig holds account handler configuration.
type AccountRouteConfig struct {
	UserService *user.UserService
	Auth        *auth.AuthService
}

// NewEngine creates a new Gin engine with HTML templates and optional structured logging.
// If logger is nil, request logging middleware is not added.
func NewEngine(logger log.Logger) *gin.Engine {
	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(RequestIDMiddleware())
	if logger != nil {
		e.Use(ZapLoggerMiddleware(logger))
	}
	e.LoadHTMLGlob(filepath.Join("internal", "server", "http", "templates", "*.html"))
	return e
}

// ZapLoggerMiddleware logs each request with method, path, status, request_id, and latency.
func ZapLoggerMiddleware(l log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		c.Next()
		status := c.Writer.Status()
		latency := time.Since(start)
		reqID := GetRequestID(c)
		fields := []zap.Field{
			zap.String("request_id", reqID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		}
		if status >= 500 {
			l.Error("request", fields...)
		} else if status >= 400 {
			l.Warn("request", fields...)
		} else {
			l.Info("request", fields...)
		}
	}
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

// RegisterHealthRoutes adds health check endpoints to the given engine.
func RegisterHealthRoutes(e *gin.Engine, cfg *HealthRouteConfig) {
	if cfg == nil {
		return
	}
	h := NewHealthHandler(cfg.Client)
	e.GET("/healthz", h.Healthz)
	e.GET("/ready", h.Ready)
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

// RegisterAccountRoutes adds account management endpoints (e.g. delete account).
func RegisterAccountRoutes(e *gin.Engine, cfg *AccountRouteConfig) {
	if cfg == nil || cfg.UserService == nil || cfg.Auth == nil {
		return
	}
	h := NewAccountHandler(cfg.UserService, cfg.Auth)
	e.GET("/account/delete", h.DeleteGet)
	e.POST("/account/delete", h.DeletePost)
}
