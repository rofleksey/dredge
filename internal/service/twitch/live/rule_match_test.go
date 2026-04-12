package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestParseLoginSet_starMeansAll(t *testing.T) {
	t.Parallel()

	all, set := parseLoginSet("*", true)
	assert.True(t, all)
	assert.Nil(t, set)

	all, set = parseLoginSet("  *  ", true)
	assert.True(t, all)
	assert.Nil(t, set)
}

func TestParseLoginSet_emptyNotStar(t *testing.T) {
	t.Parallel()

	all, set := parseLoginSet("", false)
	assert.True(t, all)
	assert.Nil(t, set)
}

func TestParseLoginSet_commaSeparated(t *testing.T) {
	t.Parallel()

	all, set := parseLoginSet(" A , b ", true)
	assert.False(t, all)
	require.Len(t, set, 2)
	_, okA := set["a"]
	_, okB := set["b"]

	assert.True(t, okA)
	assert.True(t, okB)
}

func TestCompileRules_invalidRegex(t *testing.T) {
	t.Parallel()

	rules := []entity.Rule{{Regex: "("}}
	out, errs := compileRules(rules)
	assert.Empty(t, out)
	require.Len(t, errs, 1)
}

func TestCompileRules_ok(t *testing.T) {
	t.Parallel()

	rules := []entity.Rule{{
		ID:               1,
		Regex:            `hello`,
		IncludedUsers:    "*",
		DeniedUsers:      "bad",
		IncludedChannels: "c1",
		DeniedChannels:   "",
	}}
	out, errs := compileRules(rules)
	require.Empty(t, errs)
	require.Len(t, out, 1)
	assert.Equal(t, int64(1), out[0].entity.ID)
	require.NotNil(t, out[0].re)
}

func TestCompiledRule_matches(t *testing.T) {
	t.Parallel()

	rules := []entity.Rule{{
		Regex:            `^ping$`,
		IncludedUsers:    "alice",
		DeniedUsers:      "bob",
		IncludedChannels: "chan1",
		DeniedChannels:   "blocked",
	}}
	out, errs := compileRules(rules)
	require.Empty(t, errs)
	require.Len(t, out, 1)

	cr := &out[0]

	assert.False(t, cr.matches("other", "chan1", "ping"), "user not allowed")
	assert.False(t, cr.matches("bob", "chan1", "ping"), "denied user")
	assert.False(t, cr.matches("alice", "other", "ping"), "channel not in include set")
	assert.False(t, cr.matches("alice", "blocked", "ping"), "denied channel")
	assert.False(t, cr.matches("alice", "chan1", "pong"), "regex mismatch")
	assert.True(t, cr.matches("alice", "chan1", "ping"))
}
