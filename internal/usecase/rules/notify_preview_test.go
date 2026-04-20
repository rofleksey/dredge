package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_notifyDisplayTextForLog_expandedPreferred(t *testing.T) {
	t.Parallel()

	p := EvalPayload{Event: EventStreamStart, Channel: "foo", Title: "T"}
	got := notifyDisplayTextForLog(p, "custom line")

	assert.Equal(t, "custom line", got)
}

func Test_notifyDisplayTextForLog_streamStartDefault(t *testing.T) {
	t.Parallel()

	p := EvalPayload{Event: EventStreamStart, Channel: "foo", Title: "My title"}
	got := notifyDisplayTextForLog(p, "")

	assert.Equal(t, "[live] #foo started streaming: My title", got)
}

func Test_notifyDisplayTextForLog_streamEndDefault(t *testing.T) {
	t.Parallel()

	p := EvalPayload{Event: EventStreamEnd, Channel: "bar"}
	got := notifyDisplayTextForLog(p, "")

	assert.Equal(t, "[offline] #bar stopped streaming", got)
}

func Test_notifyDisplayTextForLog_chatDefault(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", telegramChatMsgTruncateRunes+10)
	p := EvalPayload{Event: EventChatMessage, Channel: "c", Username: "u", Text: long}
	got := notifyDisplayTextForLog(p, "")

	prefix := "[c] u: "
	assert.Len(t, []rune(got), len([]rune(prefix))+telegramChatMsgTruncateRunes)
}
