package live

import (
	"testing"

	"github.com/rofleksey/dredge/internal/entity"
)

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
