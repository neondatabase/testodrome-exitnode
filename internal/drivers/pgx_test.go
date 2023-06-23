package drivers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/petuhovskiy/neon-lights/internal/log"
	"github.com/petuhovskiy/neon-lights/internal/models"
)

type dummyQuerySaver struct{}

func (dummyQuerySaver) Save(query *models.Query) error {
	log.Info(context.Background(), "save query", zap.Any("query", query))
	return nil
}

func Test_PgxConnection(t *testing.T) {
	log.DefaultGlobals()

	const exampleConnstr = "postgres://user:pass@ep-abc-xyz-123.eu-west-1.aws.neon.build/neondb"
	srv, err := PgxConnect(context.Background(), exampleConnstr, dummyQuerySaver{})
	assert.Error(t, err)

	if err == nil {
		_, err = srv.Query(context.Background(), SingleQuery{
			Query:  "SELECT 1",
			Params: nil,
		})
		assert.NoError(t, err)
	}
}
