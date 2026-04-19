package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_ListTwitchMessages(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListChatMessages(gomock.Any(), gomock.Any()).Return([]entity.ChatHistoryMessage{
		{ID: 1, Channel: "c", Username: "u", Message: "hi", MsgType: "irc"},
	}, nil)

	out, err := h.ListTwitchMessages(context.Background(), gen.ListTwitchMessagesParams{})
	require.NoError(t, err)
	require.Len(t, out, 1)
}
