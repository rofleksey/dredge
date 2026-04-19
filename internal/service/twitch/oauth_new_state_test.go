package twitch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth_NewState_VerifyState(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	state, err := o.NewState("")
	require.NoError(t, err)
	require.NotEmpty(t, state)

	_, err = o.VerifyState(state)
	require.NoError(t, err)
}

func TestOAuth_VerifyState_rejectsReuse(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	state, err := o.NewState("")
	require.NoError(t, err)

	_, err = o.VerifyState(state)
	require.NoError(t, err)
	_, err = o.VerifyState(state)
	assert.Error(t, err)
}

func TestOAuth_VerifyState_errors(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	_, err := o.VerifyState("bad")
	assert.Error(t, err)
	_, err = o.VerifyState("a.b.c")
	assert.Error(t, err)
}

func TestOAuth_VerifyState_signatureMismatch(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	st, err := o.NewState("")
	require.NoError(t, err)

	parts := strings.Split(st, ".")
	require.Len(t, parts, 2)

	tampered := parts[0] + ".YmFk" // "bad" base64 raw

	_, err = o.VerifyState(tampered)
	assert.Error(t, err)
}
