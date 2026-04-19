package twitch

import (
	"strings"

	"github.com/rofleksey/dredge/internal/entity"
)

// PatchTouchesSuspicionFields reports whether a settings patch can change suspicion state shown in the UI.
func PatchTouchesSuspicionFields(p entity.TwitchUserPatch) bool {
	return p.IsSus != nil || p.SusType != nil || p.SusDescription != nil || p.SusAutoSuppressed != nil
}

// BroadcastTwitchUserSuspicion pushes current suspicion fields to all live WebSocket clients.
func (s *Usecase) BroadcastTwitchUserSuspicion(u entity.TwitchUser) {
	if s == nil || s.broadcaster == nil {
		return
	}

	login := strings.ToLower(strings.TrimSpace(u.Username))

	payload := map[string]any{
		"type":           "twitch_user_suspicion",
		"user_twitch_id": u.ID,
		"username":       login,
		"is_sus":         u.IsSus,
	}

	if u.SusType != nil {
		st := strings.TrimSpace(*u.SusType)
		if st != "" {
			payload["sus_type"] = st
		}
	}

	if u.SusDescription != nil {
		sd := strings.TrimSpace(*u.SusDescription)
		if sd != "" {
			payload["sus_description"] = sd
		}
	}

	s.broadcaster.BroadcastJSON(payload)
}
