package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_ListChatHistory_ok(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "chan").Return(true, nil)
	repo.EXPECT().ListChatHistory(gomock.Any(), "chan", 50).Return([]entity.ChatHistoryMessage{
		{ID: 1, Channel: "chan", Username: "u", Message: "m", MsgType: "irc"},
	}, nil)

	res, err := h.ListChatHistory(context.Background(), gen.ListChatHistoryParams{Channel: "chan"})
	require.NoError(t, err)

	_, ok := res.(*gen.ListChatHistoryOKApplicationJSON)
	require.True(t, ok)
}

func TestHandler_ListChatHistory_notMonitored(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "x").Return(false, nil)

	res, err := h.ListChatHistory(context.Background(), gen.ListChatHistoryParams{Channel: "x"})
	require.NoError(t, err)

	_, ok := res.(*gen.ErrorMessage)
	require.True(t, ok)
}
