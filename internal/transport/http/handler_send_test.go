package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_SendMessage_accountNotFound(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchAccountByID(gomock.Any(), int64(99)).Return(entity.TwitchAccount{}, entity.ErrTwitchAccountNotFound)

	req := &gen.SendMessageRequest{}
	req.SetAccountID(99)
	req.SetChannel("chan")
	req.SetMessage("hi")

	res, err := h.SendMessage(adminCtx(), req)
	require.NoError(t, err)

	_, ok := res.(*gen.SendMessageUnprocessableEntity)
	require.True(t, ok)
}

func TestHandler_SendMessage_forbidden(t *testing.T) {
	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()

	req := &gen.SendMessageRequest{}
	req.SetAccountID(1)
	req.SetChannel("c")
	req.SetMessage("m")

	ctx := context.WithValue(context.WithValue(context.Background(), userIDCtxKey, int64(1)), roleCtxKey, "user")

	_, err := h.SendMessage(ctx, req)
	require.Error(t, err)
}
