package domain

// IdPConnector holds configuration for a federated upstream identity provider.
type IdPConnector struct {
	ID           string
	Issuer       string
	ClientID     string
	ClientSecret string
}
