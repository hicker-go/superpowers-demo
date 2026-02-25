package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/qinzj/superpowers-demo/ent"
	"github.com/qinzj/superpowers-demo/internal/domain"
)

// IdPConnectorRepository implements federation.IdPConnectorRepository using ent.
type IdPConnectorRepository struct {
	client *ent.Client
}

// NewIdPConnectorRepository creates an IdPConnectorRepository backed by the given ent client.
func NewIdPConnectorRepository(client *ent.Client) *IdPConnectorRepository {
	return &IdPConnectorRepository{client: client}
}

// GetByID returns the IdPConnector with the given ID, or nil if not found.
// ID can be ent's numeric ID as string (e.g. "1").
func (r *IdPConnectorRepository) GetByID(ctx context.Context, id string) (*domain.IdPConnector, error) {
	numericID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connector id: %w", err)
	}
	entConn, err := r.client.IdPConnector.Get(ctx, numericID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get idp connector: %w", err)
	}
	return entIdPConnectorToDomain(entConn), nil
}

// List returns all IdP connectors.
func (r *IdPConnectorRepository) List(ctx context.Context) ([]*domain.IdPConnector, error) {
	ents, err := r.client.IdPConnector.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list idp connectors: %w", err)
	}
	out := make([]*domain.IdPConnector, len(ents))
	for i, e := range ents {
		out[i] = entIdPConnectorToDomain(e)
	}
	return out, nil
}

func entIdPConnectorToDomain(e *ent.IdPConnector) *domain.IdPConnector {
	return &domain.IdPConnector{
		ID:           strconv.Itoa(e.ID),
		Issuer:       e.Issuer,
		ClientID:     e.ClientID,
		ClientSecret: e.ClientSecret,
	}
}
