package twitch

import (
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

var (
	ErrInvalidChannelName   = helix.ErrInvalidChannelName
	ErrUnknownTwitchChannel = helix.ErrUnknownTwitchChannel
	// ErrSendChatTimeout is returned when the Helix chat send does not complete before the deadline.
	ErrSendChatTimeout = helix.ErrSendChatTimeout
	// ErrChannelNotMonitored is returned when chat history is requested for a channel not in settings.
	ErrChannelNotMonitored = errors.New("channel is not monitored")
	// ErrNoLinkedTwitchAccount is returned when OAuth is required but no Twitch account is linked.
	ErrNoLinkedTwitchAccount = entity.ErrNoLinkedTwitchAccount
)

// SendChatNoticeError is returned when Twitch rejects the message via the Helix Chat API.
type SendChatNoticeError = helix.ChatSendError
