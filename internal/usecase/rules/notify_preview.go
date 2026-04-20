package rules

import (
	"fmt"
	"strings"
)

const (
	telegramChatMsgTruncateRunes = 3500
	telegramStreamTitleTruncate  = 500
)

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}

	r := []rune(s)
	if len(r) <= max {
		return s
	}

	return string(r[:max])
}

// notifyDisplayTextForLog returns the outbound line stored for the rule triggers feed.
// When expandedTemplate is non-empty it matches what Telegram receives from the rules engine.
// When empty, defaults mirror internal/service/twitch/live/notify.go sendTelegram* helpers.
func notifyDisplayTextForLog(p EvalPayload, expandedTemplate string) string {
	if strings.TrimSpace(expandedTemplate) != "" {
		return expandedTemplate
	}

	switch p.Event {
	case EventStreamStart:
		text := fmt.Sprintf("[live] #%s started streaming", p.Channel)
		if strings.TrimSpace(p.Title) != "" {
			text += ": " + truncateRunes(p.Title, telegramStreamTitleTruncate)
		}

		return text
	case EventStreamEnd:
		return fmt.Sprintf("[offline] #%s stopped streaming", p.Channel)
	case EventChatMessage:
		return fmt.Sprintf("[%s] %s: %s", p.Channel, p.Username, truncateRunes(p.Text, telegramChatMsgTruncateRunes))
	default:
		return expandedTemplate
	}
}
