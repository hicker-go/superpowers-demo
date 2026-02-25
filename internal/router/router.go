// Package router provides HTTP routing definitions.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/server/http/handler"
)

// Config holds route registration configuration.
type Config struct {
	OIDC *handler.OIDCRouteConfig
}

// Setup registers all routes on the given engine.
func Setup(e *gin.Engine, cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.OIDC != nil {
		handler.RegisterOIDCRoutes(e, cfg.OIDC)
	}
}
