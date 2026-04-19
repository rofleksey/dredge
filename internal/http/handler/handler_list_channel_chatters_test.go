package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_ListChannelChatters_accountNotFound(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchAccountByID(gomock.Any(), int64(5)).Return(entity.TwitchAccount{}, entity.ErrTwitchAccountNotFound)

	req := &gen.ListChannelChattersRequest{}
	req.SetAccountID(5)
	req.SetLogin("channel")

	res, err := h.ListChannelChatters(context.Background(), req)
	require.NoError(t, err)

	_, ok := res.(*gen.ErrorMessage)
	require.True(t, ok)
}
