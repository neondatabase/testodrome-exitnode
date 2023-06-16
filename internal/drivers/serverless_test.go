package drivers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/petuhovskiy/neon-lights/internal/models"
)

func Test_NewServerless(t *testing.T) {
	const exampleConnstr = "postgres://user:pass@ep-abc-xyz-123.eu-west-1.aws.neon.build/neondb"
	srv, err := NewServerless("", &models.Project{
		ConnectionString: exampleConnstr,
	})
	assert.NoError(t, err)

	assert.Equal(t, "https://ep-abc-xyz-123.eu-west-1.aws.neon.build/sql", srv.httpURL())
	assert.Equal(t, exampleConnstr, srv.connstr.String())
}
