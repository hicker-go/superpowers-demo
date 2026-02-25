// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package handler

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/qinzj/superpowers-demo/internal/service/auth"
)

const sessionCookieName = "sso_session"
const sessionCookieMaxAge = 86400 // 24 hours in seconds

// LoginHandler handles the login page and form submission.
type LoginHandler struct {
	Auth *auth.AuthService
}

// NewLoginHandler creates a LoginHandler with the given auth service.
func NewLoginHandler(a *auth.AuthService) *LoginHandler {
	return &LoginHandler{Auth: a}
}

// LoginParams holds the OAuth2 authorize params passed to/from the login page.
type LoginParams struct {
	ClientID     string `form:"client_id"`
	RedirectURI  string `form:"redirect_uri"`
	ResponseType string `form:"response_type"`
	Scope        string `form:"scope"`
	State        string `form:"state"`
}

// LoginForm holds the POST form fields.
type LoginForm struct {
	LoginParams
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// GetLogin renders the login page with OAuth2 params preserved as hidden fields.
func (h *LoginHandler) GetLogin(c *gin.Context) {
	params := LoginParams{
		ClientID:     c.Query("client_id"),
		RedirectURI:  c.Query("redirect_uri"),
		ResponseType: c.Query("response_type"),
		Scope:        c.Query("scope"),
		State:        c.Query("state"),
	}
	c.HTML(http.StatusOK, "login.html", params)
}

// PostLogin processes the login form, validates credentials, creates session, and redirects to /authorize.
func (h *LoginHandler) PostLogin(c *gin.Context) {
	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", loginTemplateData(form.LoginParams, "Invalid form"))
		return
	}

	ctx := c.Request.Context()
	user, err := h.Auth.ValidateCredentials(ctx, form.Username, form.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.HTML(http.StatusUnauthorized, "login.html", loginTemplateData(form.LoginParams, "Invalid username or password"))
			return
		}
		c.HTML(http.StatusInternalServerError, "login.html", loginTemplateData(form.LoginParams, "Authentication error"))
		return
	}
	if user == nil {
		c.HTML(http.StatusUnauthorized, "login.html", loginTemplateData(form.LoginParams, "Invalid username or password"))
		return
	}

	sess, err := h.Auth.CreateSession(ctx, user.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", loginTemplateData(form.LoginParams, "Session creation failed"))
		return
	}

	c.SetCookie(sessionCookieName, sess.Token, sessionCookieMaxAge, "/", "", false, true)

	authURL := buildAuthorizeURL(form.LoginParams)
	c.Redirect(http.StatusFound, authURL)
}

// loginTemplateData merges LoginParams with an optional error for template rendering.
func loginTemplateData(p LoginParams, errMsg string) gin.H {
	return gin.H{
		"ClientID":     p.ClientID,
		"RedirectURI":  p.RedirectURI,
		"ResponseType": p.ResponseType,
		"Scope":        p.Scope,
		"State":        p.State,
		"Error":        errMsg,
	}
}

func buildAuthorizeURL(p LoginParams) string {
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("response_type", p.ResponseType)
	q.Set("scope", p.Scope)
	q.Set("state", p.State)
	return "/authorize?" + q.Encode()
}
