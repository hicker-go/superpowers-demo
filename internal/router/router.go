// Package router provides HTTP routing definitions.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/server/http/handler"
)

// Config holds route registration configuration.
type Config struct {
	OIDC     *handler.OIDCRouteConfig
	Login    *handler.LoginRouteConfig
	Register *handler.RegisterRouteConfig
}

// Setup registers all routes on the given engine.
func Setup(e *gin.Engine, cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.OIDC != nil {
		handler.RegisterOIDCRoutes(e, cfg.OIDC)
	}
	if cfg.Login != nil {
		handler.RegisterLoginRoutes(e, cfg.Login)
	}
	if cfg.Register != nil {
		handler.RegisterRegisterRoutes(e, cfg.Register)
	}
}
