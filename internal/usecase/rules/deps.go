package rules

import (
	"context"
)

// NotifyDispatcher sends outbound notifications (Telegram, webhook).
type NotifyDispatcher interface {
	// NotifyChatKeyword is for chat-backed rules (keyword / message context).
	NotifyChatKeyword(ctx context.Context, channel, user, message, textTemplate string)
	// NotifyRuleText is for rules with no chat line (e.g. interval): channel + rendered template only.
	NotifyRuleText(ctx context.Context, channel, text string)
	NotifyStreamStart(ctx context.Context, channel, title, textTemplate string)
	NotifyStreamEnd(ctx context.Context, channel, textTemplate string)
}

// SendMessenger sends a Twitch chat message via Helix.
type SendMessenger interface {
	SendMessage(ctx context.Context, accountID int64, channel, message string) error
}
