package helix

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelixAPIFailure_plainBody(t *testing.T) {
	t.Parallel()

	err := helixAPIFailure(500, []byte(`not json`))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "500"))
}
