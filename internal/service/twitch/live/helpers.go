package live

import (
	"strings"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"github.com/rofleksey/dredge/internal/entity"
)

// NormalizeTwitchChannel returns a canonical lowercase channel name without #.
func NormalizeTwitchChannel(ch string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ch)), "#")
}

// badgeTagsFromIRC maps Twitch IRC user state to UI badge tags (mod / vip / bot / other).
func badgeTagsFromIRC(user twitchirc.User) []string {
	var out []string

	if user.IsMod {
		out = append(out, "moderator")
	}

	if user.IsVip {
		out = append(out, "vip")
	}

	if _, ok := user.Badges["bot"]; ok {
		out = append(out, "bot")
	}

	skip := map[string]struct{}{
		"broadcaster": {},
		"moderator":   {},
		"vip":         {},
		"bot":         {},
	}

	hasOther := false

	for name := range user.Badges {
		if _, s := skip[name]; s {
			continue
		}

		hasOther = true

		break
	}

	if hasOther {
		out = append(out, "other")
	}

	return out
}

// ircChannelJoinWanted returns whether the IRC monitor should be in the channel's chat for this user row.
func ircChannelJoinWanted(u entity.TwitchUser, helixLive bool) bool {
	if !u.Monitored {
		return false
	}

	if !u.IrcOnlyWhenLive {
		return true
	}

	return helixLive
}

func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}

	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}
