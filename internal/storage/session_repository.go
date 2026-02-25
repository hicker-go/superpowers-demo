package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/qinzj/superpowers-demo/ent"
	"github.com/qinzj/superpowers-demo/ent/session"
	"github.com/qinzj/superpowers-demo/internal/domain"
)

// SessionRepository implements auth.SessionRepository using ent.
type SessionRepository struct {
	client *ent.Client
}

// NewSessionRepository creates a SessionRepository backed by the given ent client.
func NewSessionRepository(client *ent.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

// Create persists the session. Session.ID and Token must be set; UserID is required.
func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	userID, err := strconv.Atoi(s.UserID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	_, err = r.client.Session.Create().
		SetToken(s.Token).
		SetExpiresAt(s.ExpiresAt).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetByToken returns the session with the given token if it exists and is not expired.
func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	entSession, err := r.client.Session.Query().
		Where(session.TokenEQ(token)).
		Where(session.ExpiresAtGT(time.Now())).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query session by token: %w", err)
	}
	return entSessionToDomain(entSession), nil
}

func entSessionToDomain(e *ent.Session) *domain.Session {
	s := &domain.Session{
		ID:        strconv.Itoa(e.ID),
		Token:     e.Token,
		ExpiresAt: e.ExpiresAt,
	}
	if e.Edges.User != nil {
		s.UserID = strconv.Itoa(e.Edges.User.ID)
	}
	return s
}

// GetByTokenWithUser returns the session and its user if valid and not expired.
func (r *SessionRepository) GetByTokenWithUser(ctx context.Context, token string) (*domain.Session, *domain.User, error) {
	entSession, err := r.client.Session.Query().
		Where(session.TokenEQ(token)).
		Where(session.ExpiresAtGT(time.Now())).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("query session by token: %w", err)
	}
	s := entSessionToDomain(entSession)
	var u *domain.User
	if entSession.Edges.User != nil {
		u = &domain.User{
			ID:       strconv.Itoa(entSession.Edges.User.ID),
			Username: entSession.Edges.User.Username,
			Email:    entSession.Edges.User.Email,
		}
	}
	return s, u, nil
}
