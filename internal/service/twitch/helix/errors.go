package helix

import "errors"

var (
	ErrInvalidChannelName   = errors.New("invalid twitch channel name")
	ErrInvalidGameID        = errors.New("invalid twitch game id")
	ErrUnknownTwitchChannel = errors.New("unknown twitch channel")
	// ErrHelixUpstream is returned when Helix returns a non-success response or an unexpected payload for a request that should have succeeded.
	ErrHelixUpstream = errors.New("helix upstream error")

	// ErrSendChatTimeout is returned when a Helix chat send does not finish before the deadline.
	ErrSendChatTimeout = errors.New("twitch chat send timeout")
)
