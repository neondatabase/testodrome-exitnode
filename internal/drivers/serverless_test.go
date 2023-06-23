package drivers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewServerless(t *testing.T) {
	const exampleConnstr = "postgres://user:pass@ep-abc-xyz-123.eu-west-1.aws.neon.build/neondb"
	srv, err := NewServerless(exampleConnstr, dummyQuerySaver{})
	assert.NoError(t, err)

	assert.Equal(t, "https://ep-abc-xyz-123.eu-west-1.aws.neon.build/sql", srv.httpURL())
	assert.Equal(t, exampleConnstr, srv.connstr.String())
}
