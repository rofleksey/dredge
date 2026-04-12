package twitch

import "github.com/rofleksey/dredge/internal/service/twitch/helix"

// ErrSendChatTimeout is returned when the Helix chat send does not complete before the deadline.
var ErrSendChatTimeout = helix.ErrSendChatTimeout

// SendChatNoticeError is returned when Twitch rejects the message via the Helix Chat API.
type SendChatNoticeError = helix.ChatSendError
