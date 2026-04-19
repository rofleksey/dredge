package twitch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_GetTwitchUserProfile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	now := time.Now().UTC()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(1)).Return(entity.TwitchUser{ID: 1}, nil)
	repo.EXPECT().CountChatMessagesByChatter(gomock.Any(), int64(1)).Return(int64(2), nil)
	repo.EXPECT().ListUserActivityEventsForTimeline(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().GetHelixMeta(gomock.Any(), int64(1)).Return(&now, &now, nil, nil)
	repo.EXPECT().ListFollowedMonitoredChannels(gomock.Any(), int64(1)).Return(nil, nil)
	repo.EXPECT().ListUserFollowedChannels(gomock.Any(), int64(1)).Return(nil, nil)
	repo.EXPECT().ListChannelBlacklist(gomock.Any()).Return(nil, nil)

	u, n, pres, ac, _, follows, gqlFollows, bl, err := svc.GetTwitchUserProfile(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, int64(2), n)
	assert.Equal(t, int64(0), pres)
	assert.NotNil(t, ac)
	assert.Empty(t, follows)
	assert.Empty(t, gqlFollows)
	assert.Empty(t, bl)
}
