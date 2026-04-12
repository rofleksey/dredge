package twitch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestComputeAutoSuspicion(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	accountOld := now.Add(-400 * 24 * time.Hour)

	settings := entity.SuspicionSettings{
		AutoCheckAccountAge: true,
		AccountAgeSusDays:   14,
		AutoCheckBlacklist:  true,
		AutoCheckLowFollows: true,
		LowFollowsThreshold: 10,
	}

	bl := map[string]struct{}{"badchannel": {}}

	t.Run("blacklist_wins", func(t *testing.T) {
		t.Parallel()

		follows := []entity.FollowedChannelRow{
			{FollowedChannelLogin: "badchannel"},
		}
		ok, typ, desc := computeAutoSuspicion(settings, bl, follows, 100, &accountOld, now)
		assert.True(t, ok)
		assert.Equal(t, entity.SusTypeAutoBlacklist, typ)
		assert.Contains(t, desc, "blacklist")
	})

	t.Run("age_new_account", func(t *testing.T) {
		t.Parallel()

		created := now.Add(-5 * 24 * time.Hour)
		ok, typ, _ := computeAutoSuspicion(settings, nil, nil, 100, &created, now)
		assert.True(t, ok)
		assert.Equal(t, entity.SusTypeAutoAge, typ)
	})

	t.Run("low_follows", func(t *testing.T) {
		t.Parallel()

		ok, typ, _ := computeAutoSuspicion(settings, nil, nil, 3, &accountOld, now)
		assert.True(t, ok)
		assert.Equal(t, entity.SusTypeAutoLowFollow, typ)
	})

	t.Run("clean", func(t *testing.T) {
		t.Parallel()

		ok, _, _ := computeAutoSuspicion(settings, nil, nil, 50, &accountOld, now)
		assert.False(t, ok)
	})

	t.Run("disabled_checks", func(t *testing.T) {
		t.Parallel()

		s := settings
		s.AutoCheckAccountAge = false
		s.AutoCheckBlacklist = false
		s.AutoCheckLowFollows = false
		ok, _, _ := computeAutoSuspicion(s, bl, []entity.FollowedChannelRow{{FollowedChannelLogin: "badchannel"}}, 2, nil, now)
		assert.False(t, ok)
	})
}
