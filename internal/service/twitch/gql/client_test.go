package gql

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	c := NewClient(nil)
	require.NotNil(t, c)
	require.NotNil(t, c.HTTPClient)

	c2 := NewClient(http.DefaultClient)
	require.Equal(t, http.DefaultClient, c2.HTTPClient)
}
