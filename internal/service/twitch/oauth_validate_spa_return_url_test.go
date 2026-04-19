package twitch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth_ValidateSPAReturnURL_empty(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	require.NoError(t, o.ValidateSPAReturnURL(""))
}

func TestOAuth_ValidateSPAReturnURL_mismatch(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	err := o.ValidateSPAReturnURL("https://evil.example/")
	assert.Error(t, err)
}
