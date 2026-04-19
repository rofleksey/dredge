package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_GetTwitchUserActivityTimeline(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(8)).Return(entity.TwitchUser{ID: 8}, nil)
	repo.EXPECT().ListUserActivityEventsForTimeline(gomock.Any(), int64(8), gomock.Any(), gomock.Any()).Return(nil, nil)

	req := &gen.GetTwitchUserActivityTimelineRequest{}
	req.SetID(8)

	res, err := h.GetTwitchUserActivityTimeline(context.Background(), req)
	require.NoError(t, err)

	_, ok := res.(*gen.GetTwitchUserActivityTimelineOKApplicationJSON)
	require.True(t, ok)
}
