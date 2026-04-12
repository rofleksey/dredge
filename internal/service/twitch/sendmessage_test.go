package twitch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

func TestSendChatNoticeError_Error(t *testing.T) {
	t.Parallel()

	var e *SendChatNoticeError
	assert.Equal(t, "", e.Error())

	e = &SendChatNoticeError{Message: "x"}
	assert.Equal(t, "x", e.Error())

	var h *helix.ChatSendError
	assert.ErrorAs(t, e, &h)
}
