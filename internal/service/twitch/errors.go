package twitch

import (
	"errors"

	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

var (
	ErrInvalidChannelName   = helix.ErrInvalidChannelName
	ErrUnknownTwitchChannel = helix.ErrUnknownTwitchChannel
	// ErrChannelNotMonitored is returned when chat history is requested for a channel not in settings.
	ErrChannelNotMonitored = errors.New("channel is not monitored")
	// ErrNoLinkedTwitchAccount is returned when OAuth is required but no Twitch account is linked.
	ErrNoLinkedTwitchAccount = errors.New("no linked twitch account")
)
