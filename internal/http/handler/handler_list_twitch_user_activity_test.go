package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_ListTwitchUserActivity(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(7)).Return(entity.TwitchUser{ID: 7, Username: "who"}, nil)
	repo.EXPECT().ListUserActivityEvents(gomock.Any(), gomock.Any()).Return([]entity.UserActivityEvent{
		{ID: 1, EventType: entity.UserActivityChatOnline, ChatterTwitchUserID: 7},
	}, nil)

	req := &gen.ListTwitchUserActivityRequest{}
	req.SetID(7)

	res, err := h.ListTwitchUserActivity(context.Background(), req)
	require.NoError(t, err)

	_, ok := res.(*gen.ListTwitchUserActivityOKApplicationJSON)
	require.True(t, ok)
}
