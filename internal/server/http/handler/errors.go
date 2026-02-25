// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/federation"
	"github.com/qinzj/superpowers-demo/internal/service/user"
)

const requestIDKey = "request_id"
const requestIDHeader = "X-Request-ID"

// ErrorResp is the standard error response shape (code, message, request_id).
type ErrorResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// RequestIDMiddleware sets a request_id (UUID) on the context if not present.
// Uses X-Request-ID from the request if provided; otherwise generates a new UUID.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(requestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// GetRequestID returns the request_id from gin.Context, or empty string if not set.
func GetRequestID(c *gin.Context) string {
	v, _ := c.Get(requestIDKey)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ErrorMapping maps service errors to HTTP status and error codes.
func ErrorMapping(err error) (status int, code string) {
	switch {
	case err == nil:
		return http.StatusOK, ""
	case errors.Is(err, auth.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid_credentials"
	case errors.Is(err, user.ErrUsernameTaken):
		return http.StatusConflict, "username_taken"
	case errors.Is(err, user.ErrWeakPassword):
		return http.StatusBadRequest, "weak_password"
	case errors.Is(err, federation.ErrConnectorNotFound):
		return http.StatusNotFound, "connector_not_found"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}

// WriteError maps err to HTTP status and code, then writes ErrorResp as JSON.
// Uses fallbackMessage when provided; otherwise err.Error(). For 500, message is sanitized.
func WriteError(c *gin.Context, err error, fallbackMessage string) {
	status, code := ErrorMapping(err)
	msg := fallbackMessage
	if msg == "" && err != nil {
		msg = err.Error()
	}
	if status == http.StatusInternalServerError {
		msg = "An unexpected error occurred"
	}
	writeErrorResp(c, status, code, msg)
}

// WriteErrorWithStatus writes ErrorResp with explicit status and code (for non-service errors).
func WriteErrorWithStatus(c *gin.Context, status int, code, message string) {
	writeErrorResp(c, status, code, message)
}

func writeErrorResp(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResp{
		Code:      code,
		Message:   message,
		RequestID: GetRequestID(c),
	})
}
