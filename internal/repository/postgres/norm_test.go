package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeStoredUsername(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "foo", normalizeStoredUsername("  FOO  "))
	assert.Equal(t, "", normalizeStoredUsername("   "))
}

func TestNormalizeChannelName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "foo", normalizeChannelName("  #FOO  "))
	assert.Equal(t, "bar", normalizeChannelName("bar"))
	assert.Equal(t, "", normalizeChannelName("   "))
}
