package dto

// ValidateCredentialsRequest holds username and password for credential validation.
// Used by login handlers when binding form or JSON input.
type ValidateCredentialsRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}
