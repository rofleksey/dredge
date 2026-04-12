package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeTwitchAccountLinkType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bot", normalizeTwitchAccountLinkType("bot"))
	assert.Equal(t, "main", normalizeTwitchAccountLinkType("main"))
	assert.Equal(t, "main", normalizeTwitchAccountLinkType(""))
	assert.Equal(t, "main", normalizeTwitchAccountLinkType("anything"))
}
