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

func TestMwContainsWordCaseInsensitive(t *testing.T) {
	t.Parallel()

	s := map[string]any{
		"words": []any{"Foo"},
	}
	require.True(t, mwContainsWord(s, "hello foo there"))
	require.False(t, mwContainsWord(s, "hello bar"))

	sExplicit := map[string]any{
		"words":            []any{"Foo"},
		"case_insensitive": true,
	}
	require.True(t, mwContainsWord(sExplicit, "hello foo there"))

	sSensitive := map[string]any{
		"words":            []any{"Foo"},
		"case_insensitive": false,
	}
	require.True(t, mwContainsWord(sSensitive, "x Foo y"))
	require.False(t, mwContainsWord(sSensitive, "x foo y"))
}
