package twitch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth_ReturnURL(t *testing.T) {
	t.Parallel()

	o := NewOAuth("c", "s", "http://a", "http://return/here", "sixteen-byte-key!!")
	assert.Equal(t, "http://return/here", o.ReturnURL())
}
