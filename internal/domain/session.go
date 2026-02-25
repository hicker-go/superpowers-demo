package domain

import "time"

// Session represents an authenticated user session.
type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
}
