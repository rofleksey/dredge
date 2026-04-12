package live

import (
	"testing"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/stretchr/testify/assert"
)

func TestBadgeTagsFromIRC(t *testing.T) {
	t.Parallel()

	t.Run("mod_vip_bot_other", func(t *testing.T) {
		t.Parallel()

		u := twitchirc.User{
			IsMod:  true,
			IsVip:  true,
			Badges: map[string]int{"bot": 1, "subscriber": 3},
		}
		tags := badgeTagsFromIRC(u)
		assert.Equal(t, []string{"moderator", "vip", "bot", "other"}, tags)
	})

	t.Run("broadcaster_skipped_for_other", func(t *testing.T) {
		t.Parallel()

		u := twitchirc.User{
			Badges: map[string]int{"broadcaster": 1},
		}
		tags := badgeTagsFromIRC(u)
		assert.Empty(t, tags)
	})
}
