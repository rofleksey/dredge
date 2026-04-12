package live

import twitchirc "github.com/gempir/go-twitch-irc/v4"

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
