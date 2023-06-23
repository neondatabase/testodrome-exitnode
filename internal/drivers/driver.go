package drivers

import (
	"context"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

type Driver interface {
	Query(ctx context.Context, req SingleQuery) (*models.Query, error)
}

type ManyQueriesDriver interface {
	Queries(ctx context.Context, queries ...SingleQuery) ([]models.Query, error)
}

type CloseableDriver interface {
	Close(context.Context) error
}
