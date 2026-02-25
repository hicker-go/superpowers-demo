// Copyright © 2026 qinzj
// SPDX-License-Identifier: MIT

package dto

// RegisterRequest holds the registration form data.
type RegisterRequest struct {
	Username string `form:"username" binding:"required,min=1,max=64"`
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required,min=8"`
}
