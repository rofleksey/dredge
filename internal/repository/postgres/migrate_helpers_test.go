package postgres

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuoteSQLString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "'plain'", quoteSQLString("plain"))
	assert.Equal(t, "'o''clock'", quoteSQLString("o'clock"))
}

func TestListMigrationFiles(t *testing.T) {
	t.Parallel()

	names, err := listMigrationFiles()
	require.NoError(t, err)
	require.Len(t, names, 2)
	assert.Equal(t, "0001_init.sql", names[0])
	assert.Equal(t, "0002_streams_viewer_count.sql", names[1])

	for _, n := range names {
		assert.True(t, strings.HasSuffix(n, ".sql"), n)
	}

	sorted := append([]string(nil), names...)
	slices.Sort(sorted)
	assert.Equal(t, sorted, names)
}
