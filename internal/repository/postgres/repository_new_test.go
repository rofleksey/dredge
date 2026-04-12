package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	r := New(nil, nil)
	assert.NotNil(t, r)
	assert.Nil(t, r.pool)
	assert.Nil(t, r.obs)
}
