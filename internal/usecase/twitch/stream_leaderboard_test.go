package twitch

import (
	"testing"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestFilterLeaderboardByQuery(t *testing.T) {
	t.Parallel()

	rows := []entity.StreamLeaderboardRow{{Login: "Alpha"}, {Login: "beta"}}
	out := filterLeaderboardByQuery(rows, "alp")
	assert.Len(t, out, 1)
	assert.Equal(t, "Alpha", out[0].Login)
}
