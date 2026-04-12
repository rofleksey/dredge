package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeTwitchChannel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "foo", NormalizeTwitchChannel("  #FOO "))
	assert.Equal(t, "bar", NormalizeTwitchChannel("bar"))
}

func TestTruncateString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", truncateString("hello", 0))
	assert.Equal(t, "", truncateString("hello", -1))
	assert.Equal(t, "hello", truncateString("hello", 10))
	assert.Equal(t, "аб...", truncateString("абвг", 2))
	assert.Equal(t, "hello...", truncateString("hello world", 5))
}
