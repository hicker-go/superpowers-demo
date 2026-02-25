package federation

import (
	"context"

	"github.com/qinzj/superpowers-demo/internal/domain"
)

// IdPConnectorRepository defines persistence operations for IdP connectors.
type IdPConnectorRepository interface {
	List(ctx context.Context) ([]*domain.IdPConnector, error)
	GetByID(ctx context.Context, id string) (*domain.IdPConnector, error)
}
