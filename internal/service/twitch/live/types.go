package live

import "context"

// RuleEngine is implemented by the rules use case engine (optional; nil disables automation).
type RuleEngine interface {
	HandleChatMessage(channel, user, text string)
	HandleStreamStart(channel, title string)
	HandleStreamEnd(channel string)
	KeywordMatchChat(ctx context.Context, channel, user, text string) bool
}
