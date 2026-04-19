package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rofleksey/dredge/internal/entity"
)

// ValidateRule checks event, middlewares, and action shape.
func ValidateRule(r entity.Rule) error {
	if r.EventType == "" {
		return fmt.Errorf("event_type required: %w", entity.ErrInvalidRule)
	}

	switch r.EventType {
	case EventChatMessage, EventStreamStart, EventStreamEnd:
	case EventInterval:
		sec, ok := numFromMap(r.EventSettings, "interval_seconds")
		if !ok || sec <= 0 {
			return fmt.Errorf("interval event requires positive interval_seconds: %w", entity.ErrInvalidRule)
		}

		ch, _ := r.EventSettings["channel"].(string)
		if ch == "" {
			return fmt.Errorf("interval event requires channel: %w", entity.ErrInvalidRule)
		}
	default:
		return fmt.Errorf("unknown event_type %q: %w", r.EventType, entity.ErrInvalidRule)
	}

	if r.ActionType == "" {
		return fmt.Errorf("action_type required: %w", entity.ErrInvalidRule)
	}

	switch r.ActionType {
	case ActionNotify:
	case ActionSendChat:
		aid, ok := numFromMap(r.ActionSettings, "account_id")
		if !ok || aid <= 0 {
			return fmt.Errorf("send_chat requires positive account_id: %w", entity.ErrInvalidRule)
		}

		ch, _ := r.ActionSettings["channel"].(string)
		if ch == "" {
			return fmt.Errorf("send_chat requires channel: %w", entity.ErrInvalidRule)
		}

		msg, _ := r.ActionSettings["message"].(string)
		if msg == "" {
			return fmt.Errorf("send_chat requires message template: %w", entity.ErrInvalidRule)
		}
	default:
		return fmt.Errorf("unknown action_type %q: %w", r.ActionType, entity.ErrInvalidRule)
	}

	for i, mw := range r.Middlewares {
		if mw.Type == "" {
			return fmt.Errorf("middleware[%d] type required: %w", i, entity.ErrInvalidRule)
		}

		if mw.Settings == nil {
			return fmt.Errorf("middleware[%d] settings required: %w", i, entity.ErrInvalidRule)
		}

		if err := validateMiddleware(mw.Type, mw.Settings); err != nil {
			return err
		}
	}

	return nil
}

func validateMiddleware(typ string, s map[string]any) error {
	switch typ {
	case MWFilterChannel, MWFilterUser:
		return nil
	case MWMatchRegex:
		pat, _ := s["pattern"].(string)
		if pat == "" {
			return fmt.Errorf("match_regex requires pattern: %w", entity.ErrInvalidRule)
		}

		if _, err := regexp.Compile(pat); err != nil {
			return fmt.Errorf("match_regex pattern: %v: %w", err, entity.ErrInvalidRule)
		}
	case MWContainsWord:
		if !containsWordsNonEmpty(s["words"]) {
			return fmt.Errorf("contains_word requires non-empty words: %w", entity.ErrInvalidRule)
		}
	case MWCooldown:
		sec, ok := numFromMap(s, "seconds")
		if !ok || sec <= 0 {
			return fmt.Errorf("cooldown requires positive seconds: %w", entity.ErrInvalidRule)
		}
	default:
		return fmt.Errorf("unknown middleware type %q: %w", typ, entity.ErrInvalidRule)
	}

	return nil
}

func containsWordsNonEmpty(v any) bool {
	if s, ok := v.([]string); ok && len(s) > 0 {
		return true
	}

	arr, ok := v.([]any)
	if !ok || len(arr) == 0 {
		return false
	}

	for _, x := range arr {
		if sv, ok := x.(string); ok && strings.TrimSpace(sv) != "" {
			return true
		}
	}

	return false
}

func numFromMap(m map[string]any, key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}

	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func strSliceFromAny(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(arr))

	for _, x := range arr {
		s, ok := x.(string)
		if !ok {
			continue
		}

		s = trimLower(s)
		if s != "" {
			out = append(out, s)
		}
	}

	return out
}

func trimLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
