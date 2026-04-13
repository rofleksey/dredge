package live

import "github.com/rofleksey/dredge/internal/entity"

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
