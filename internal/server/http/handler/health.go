// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/ent"
)

// HealthRouteConfig holds configuration for health check endpoints.
type HealthRouteConfig struct {
	Client *ent.Client
}

// HealthHandler handles /healthz and /ready endpoints.
type HealthHandler struct {
	client *ent.Client
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(client *ent.Client) *HealthHandler {
	return &HealthHandler{client: client}
}

// Healthz handles GET /healthz (liveness probe).
// Returns 200 with no dependency checks.
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Ready handles GET /ready (readiness probe).
// Checks DB connection; returns 503 if down.
func (h *HealthHandler) Ready(c *gin.Context) {
	if h.client == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if _, err := h.client.OAuth2Client.Query().Limit(1).IDs(ctx); err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	c.Status(http.StatusOK)
}
