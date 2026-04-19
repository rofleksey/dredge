package live

import (
	"testing"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/rofleksey/dredge/internal/entity"
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

func TestIrcChannelJoinWanted(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		u    entity.TwitchUser
		live bool
		want bool
	}{
		{
			name: "not_monitored",
			u:    entity.TwitchUser{Monitored: false, IrcOnlyWhenLive: false},
			live: true,
			want: false,
		},
		{
			name: "live_only_off_joins_always",
			u:    entity.TwitchUser{Monitored: true, IrcOnlyWhenLive: false},
			live: false,
			want: true,
		},
		{
			name: "live_only_on_and_stream_live",
			u:    entity.TwitchUser{Monitored: true, IrcOnlyWhenLive: true},
			live: true,
			want: true,
		},
		{
			name: "live_only_on_stream_offline",
			u:    entity.TwitchUser{Monitored: true, IrcOnlyWhenLive: true},
			live: false,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ircChannelJoinWanted(tc.u, tc.live); got != tc.want {
				t.Fatalf("ircChannelJoinWanted(...) = %v, want %v", got, tc.want)
			}
		})
	}
}
