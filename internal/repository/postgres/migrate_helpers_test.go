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
	require.Len(t, names, 7)
	assert.Equal(t, "0001_init.sql", names[0])
	assert.Equal(t, "0002_streams_viewer_count.sql", names[1])
	assert.Equal(t, "0003_enrichment_cooldown.sql", names[2])
	assert.Equal(t, "0004_twitch_user_irc_defaults_reset.sql", names[3])
	assert.Equal(t, "0005_rules_engine.sql", names[4])
	assert.Equal(t, "0006_rules_engine_legacy_repair.sql", names[5])
	assert.Equal(t, "0007_rule_name.sql", names[6])

	for _, n := range names {
		assert.True(t, strings.HasSuffix(n, ".sql"), n)
	}

	sorted := append([]string(nil), names...)
	slices.Sort(sorted)
	assert.Equal(t, sorted, names)
}
