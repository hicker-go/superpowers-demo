package domain

// OAuth2Client represents a registered OAuth2/OIDC client (RP).
type OAuth2Client struct {
	ID           string
	ClientID     string
	ClientSecret string
	RedirectURIs []string
}
