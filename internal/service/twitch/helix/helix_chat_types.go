package helix

// ChatSendError is returned when Twitch rejects a chat send via Helix (HTTP error or is_sent=false).
type ChatSendError struct {
	Message string
}

func (e *ChatSendError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}
