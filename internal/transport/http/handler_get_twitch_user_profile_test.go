package httptransport

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_GetTwitchUserProfile(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	now := time.Now().UTC()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(9)).Return(entity.TwitchUser{ID: 9, Username: "u"}, nil)
	repo.EXPECT().CountChatMessagesByChatter(gomock.Any(), int64(9)).Return(int64(1), nil)
	repo.EXPECT().ListUserActivityEventsForTimeline(gomock.Any(), int64(9), gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().GetHelixMeta(gomock.Any(), int64(9)).Return(&now, &now, nil, nil)
	repo.EXPECT().ListFollowedMonitoredChannels(gomock.Any(), int64(9)).Return(nil, nil)
	repo.EXPECT().ListUserFollowedChannels(gomock.Any(), int64(9)).Return(nil, nil)
	repo.EXPECT().ListChannelBlacklist(gomock.Any()).Return(nil, nil)

	res, err := h.GetTwitchUserProfile(context.Background(), &gen.GetTwitchUserProfileRequest{ID: 9})
	require.NoError(t, err)

	_, ok := res.(*gen.TwitchUserProfile)
	require.True(t, ok)
}
