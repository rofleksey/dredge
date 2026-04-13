package entity

import "errors"

// ErrNoSentry wraps errors that should not be reported to Sentry (expected client/auth failures).
var ErrNoSentry = errors.New("err_no_sentry")

// Sentinel errors returned by the repository layer for missing rows / business preconditions.
var (
	ErrRuleNotFound           = errors.New("rule not found")
	ErrNotificationNotFound   = errors.New("notification not found")
	ErrTwitchAccountNotFound  = errors.New("twitch account not found")
	ErrTwitchUserNotFound     = errors.New("twitch user not found")
	ErrNoTwitchUserForChannel = errors.New("unknown twitch user for channel")
	ErrStreamNotFound         = errors.New("stream not found")
	// ErrNoLinkedTwitchAccount is returned when OAuth is required but no Twitch account is linked.
	ErrNoLinkedTwitchAccount = errors.New("no linked twitch account")
	// ErrInvalidTwitchUserMonitorSettings is returned when notify_off_stream_messages is enabled without irc_only_when_live.
	ErrInvalidTwitchUserMonitorSettings = errors.New("notify_off_stream_messages requires irc_only_when_live")
)
