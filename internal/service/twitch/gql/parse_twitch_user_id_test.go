package gql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTwitchUserID(t *testing.T) {
	t.Parallel()

	n, err := parseTwitchUserID("123456789")
	require.NoError(t, err)
	assert.Equal(t, int64(123456789), n)

	n2, err := parseTwitchUserID("U123456789")
	require.NoError(t, err)
	assert.Equal(t, int64(123456789), n2)

	_, err = parseTwitchUserID("")
	assert.Error(t, err)
}
