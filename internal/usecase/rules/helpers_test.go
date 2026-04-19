package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandTemplate(t *testing.T) {
	t.Parallel()

	out := ExpandTemplate("hello $CHANNEL", map[string]string{"channel": "foo"})
	require.Equal(t, "hello foo", out)
}
