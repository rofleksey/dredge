package rules

// Event types (stored in rules.event_type).
const (
	EventChatMessage = "chat_message"
	EventStreamStart = "stream_start"
	EventStreamEnd   = "stream_end"
	EventInterval    = "interval"
)

// Middleware types.
const (
	MWFilterChannel = "filter_channel"
	MWFilterUser    = "filter_user"
	MWMatchRegex    = "match_regex"
	MWContainsWord  = "contains_word"
	MWCooldown      = "cooldown"
)

// Action types.
const (
	ActionNotify   = "notify"
	ActionSendChat = "send_chat"
)

// maxRegexRunes limits regex input size (ReDoS mitigation), same idea as live.rule_match.
const maxRegexRunes = 4000
