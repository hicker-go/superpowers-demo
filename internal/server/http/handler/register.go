// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/server/http/handler/dto"
	"github.com/qinzj/superpowers-demo/internal/service/user"
)

// RegisterHandler handles user registration.
type RegisterHandler struct {
	UserService *user.UserService
}

// NewRegisterHandler creates a RegisterHandler with the given UserService.
func NewRegisterHandler(svc *user.UserService) *RegisterHandler {
	return &RegisterHandler{UserService: svc}
}

// RegisterGet renders the registration form.
func (h *RegisterHandler) RegisterGet(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(registerFormHTML("", "", "")))
}

// RegisterPost processes the registration form.
func (h *RegisterHandler) RegisterPost(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8",
			[]byte(registerFormHTML("Invalid input", req.Username, req.Email)))
		return
	}

	_, err := h.UserService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrUsernameTaken) {
			c.Data(http.StatusConflict, "text/html; charset=utf-8",
				[]byte(registerFormHTML("Username already taken", req.Username, req.Email)))
			return
		}
		if errors.Is(err, user.ErrWeakPassword) {
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8",
				[]byte(registerFormHTML("Password must be at least 8 characters", req.Username, req.Email)))
			return
		}
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8",
			[]byte(registerFormHTML("Registration failed", req.Username, req.Email)))
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func registerFormHTML(errMsg, username, email string) string {
	errBlock := ""
	if errMsg != "" {
		errBlock = fmt.Sprintf(`<p style="color:red;">%s</p>`, html.EscapeString(errMsg))
	}
	userVal := html.EscapeString(username)
	emailVal := html.EscapeString(email)
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Register</title></head>
<body>
	<h1>Register</h1>
	%s
	<form method="POST" action="/register">
		<label>Username: <input name="username" value="%s" required maxlength="64"></label><br>
		<label>Email: <input name="email" type="email" value="%s" required></label><br>
		<label>Password: <input name="password" type="password" required minlength="8"></label><br>
		<button type="submit">Register</button>
	</form>
	<p><a href="/login">Back to Login</a></p>
</body>
</html>`, errBlock, userVal, emailVal)
}
