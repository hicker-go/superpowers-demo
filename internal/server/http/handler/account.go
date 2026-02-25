// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"fmt"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/domain"
	"github.com/qinzj/superpowers-demo/internal/service/auth"
	"github.com/qinzj/superpowers-demo/internal/service/user"
)

// AccountHandler handles account management (e.g. delete account).
type AccountHandler struct {
	UserService *user.UserService
	Auth        *auth.AuthService
}

// NewAccountHandler creates an AccountHandler with the given services.
func NewAccountHandler(userSvc *user.UserService, authSvc *auth.AuthService) *AccountHandler {
	return &AccountHandler{UserService: userSvc, Auth: authSvc}
}

// DeleteGet renders the account deletion confirmation page. Requires login.
func (h *AccountHandler) DeleteGet(c *gin.Context) {
	u := currentUser(c, h.Auth)
	if u == nil {
		c.Redirect(http.StatusFound, "/login?next=/account/delete")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(deleteAccountFormHTML(u, "")))
}

// DeletePost processes account deletion. Requires login and confirmation.
func (h *AccountHandler) DeletePost(c *gin.Context) {
	u := currentUser(c, h.Auth)
	if u == nil {
		c.Redirect(http.StatusFound, "/login?next=/account/delete")
		return
	}

	confirm := c.PostForm("confirm")
	if confirm != "yes" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8",
			[]byte(deleteAccountFormHTML(u, "Please confirm by typing 'yes'")))
		return
	}

	if err := h.UserService.Delete(c.Request.Context(), u.ID); err != nil {
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8",
			[]byte(deleteAccountFormHTML(u, "Failed to delete account. Please try again.")))
		return
	}

	// Clear session cookie
	c.SetCookie(sessionCookieName, "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

// currentUser returns the logged-in user from session, or nil.
func currentUser(c *gin.Context, authSvc *auth.AuthService) *domain.User {
	if authSvc == nil {
		return nil
	}
	token, _ := c.Cookie(sessionCookieName)
	if token == "" {
		return nil
	}
	u, err := authSvc.GetSession(c.Request.Context(), token)
	if err != nil || u == nil {
		return nil
	}
	return u
}

func deleteAccountFormHTML(u *domain.User, errMsg string) string {
	errBlock := ""
	if errMsg != "" {
		errBlock = fmt.Sprintf(`<p style="color:red;">%s</p>`, html.EscapeString(errMsg))
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Delete Account</title></head>
<body>
	<h1>Delete Account</h1>
	%s
	<p>You are about to permanently delete your account <strong>%s</strong>.</p>
	<p>This action cannot be undone. All your data will be removed.</p>
	<form method="POST" action="/account/delete">
		<label>Type <strong>yes</strong> to confirm: <input name="confirm" required></label><br>
		<button type="submit">Delete Account</button>
	</form>
	<p><a href="/login">Cancel</a></p>
</body>
</html>`, errBlock, html.EscapeString(u.Username))
}
