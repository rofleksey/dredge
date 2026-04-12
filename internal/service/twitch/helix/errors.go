package helix

import "errors"

var (
	ErrInvalidChannelName   = errors.New("invalid twitch channel name")
	ErrUnknownTwitchChannel = errors.New("unknown twitch channel")

	// ErrSendChatTimeout is returned when a Helix chat send does not finish before the deadline.
	ErrSendChatTimeout = errors.New("twitch chat send timeout")
)
