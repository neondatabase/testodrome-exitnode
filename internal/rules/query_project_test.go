package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_appendPoolerSuffix(t *testing.T) {
	res, err := appendPoolerSuffix("postgres://user:pass@ep-abc-xyz-123.eu-west-1.aws.neon.build/neondb")
	assert.NoError(t, err)
	assert.Equal(
		t,
		"postgres://user:pass@ep-abc-xyz-123-pooler.eu-west-1.aws.neon.build/neondb",
		res,
	)
}
