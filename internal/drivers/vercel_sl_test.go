package drivers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/petuhovskiy/neon-lights/internal/log"
)

func TestVercelSL(t *testing.T) {
	log.DefaultGlobals()

	cli := &VercelSL{
		connstr: "postgres://user:pass@ep-non-existent-123456.us-east-2.aws.neon.build/neondb",
		apiURL:  defaultAPIURL,
	}
	ctx := context.Background()
	_, err := cli.Queries(ctx, SingleQuery{
		Query:  "SELECT 1",
		Params: nil,
	})
	assert.NoError(t, err)
}
