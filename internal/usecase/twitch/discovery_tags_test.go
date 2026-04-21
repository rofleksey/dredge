package twitch

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamTagsCoverRequired_emptyRequired(t *testing.T) {
	t.Parallel()

	require.True(t, streamTagsCoverRequired(nil, []string{"a"}))
	require.True(t, streamTagsCoverRequired([]string{}, []string{}))
}

func TestStreamTagsCoverRequired_andSemantics(t *testing.T) {
	t.Parallel()

	require.True(t, streamTagsCoverRequired([]string{"English"}, []string{"english", "FPS"}))
	require.False(t, streamTagsCoverRequired([]string{"English", "RPG"}, []string{"english", "FPS"}))
	require.True(t, streamTagsCoverRequired([]string{" English ", "FPS"}, []string{"fps", "english"}))
}

func TestStreamTagsCoverRequired_skipsEmptyRequiredEntries(t *testing.T) {
	t.Parallel()

	require.True(t, streamTagsCoverRequired([]string{"", "  ", "x"}, []string{"X"}))
}
