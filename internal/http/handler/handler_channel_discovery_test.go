package handler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_GetChannelDiscoverySettings_smoke(t *testing.T) {
	t.Parallel()

	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetChannelDiscoverySettings(gomock.Any()).Return(entity.ChannelDiscoverySettings{
		Enabled:              false,
		PollIntervalSeconds:  3600,
		GameID:               "",
		MinLiveViewers:       0,
		RequiredStreamTags:   nil,
		MaxStreamPagesPerRun: 20,
	}, nil)

	out, err := h.GetChannelDiscoverySettings(context.Background())
	require.NoError(t, err)
	require.NotNil(t, out)
	require.False(t, out.Enabled)
}

func TestHandler_UpdateChannelDiscoverySettings_invalid(t *testing.T) {
	t.Parallel()

	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()

	res, err := h.UpdateChannelDiscoverySettings(context.Background(), &gen.ChannelDiscoverySettings{
		Enabled:              true,
		PollIntervalSeconds:  60,
		GameID:               "",
		MinLiveViewers:       0,
		RequiredStreamTags:   nil,
		MaxStreamPagesPerRun: 20,
	})
	require.NoError(t, err)
	em, ok := res.(*gen.ErrorMessage)
	require.True(t, ok)
	require.Contains(t, em.Message, "game_id")
}

func TestHandler_ListChannelDiscoveryCandidates_smoke(t *testing.T) {
	t.Parallel()

	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListTwitchDiscoveryCandidates(gomock.Any()).Return([]entity.TwitchDiscoveryCandidate{
		{
			User: entity.TwitchUser{
				ID:        1,
				Username:  "a",
				Monitored: false,
			},
			DiscoveredAt: time.Unix(100, 0).UTC(),
			LastSeenAt:   time.Unix(200, 0).UTC(),
			StreamTags:   []string{"x"},
		},
	}, nil)

	list, err := h.ListChannelDiscoveryCandidates(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, int64(1), list[0].GetUser().ID)
}

func TestHandler_ApproveChannelDiscoveryCandidate_notFound(t *testing.T) {
	t.Parallel()

	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(9)).Return(entity.TwitchUser{ID: 9, Username: "x"}, nil)
	repo.EXPECT().ApproveDiscoveryCandidate(gomock.Any(), int64(9)).Return(entity.TwitchUser{}, entity.ErrDiscoveryCandidateNotFound)

	res, err := h.ApproveChannelDiscoveryCandidate(context.Background(), gen.ApproveChannelDiscoveryCandidateParams{TwitchUserID: 9})
	require.NoError(t, err)
	_, ok := res.(*gen.ApproveChannelDiscoveryCandidateNotFound)
	require.True(t, ok)
}

func TestHandler_DenyChannelDiscoveryCandidate_noContent(t *testing.T) {
	t.Parallel()

	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().DenyDiscoveryCandidate(gomock.Any(), int64(3)).Return(nil)

	res, err := h.DenyChannelDiscoveryCandidate(context.Background(), gen.DenyChannelDiscoveryCandidateParams{TwitchUserID: 3})
	require.NoError(t, err)
	_, ok := res.(*gen.DenyChannelDiscoveryCandidateNoContent)
	require.True(t, ok)
}

func TestHandler_DenyChannelDiscoveryCandidate_notFound(t *testing.T) {
	t.Parallel()

	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().DenyDiscoveryCandidate(gomock.Any(), int64(4)).Return(entity.ErrDiscoveryCandidateNotFound)

	res, err := h.DenyChannelDiscoveryCandidate(context.Background(), gen.DenyChannelDiscoveryCandidateParams{TwitchUserID: 4})
	require.NoError(t, err)
	em, ok := res.(*gen.ErrorMessage)
	require.True(t, ok)
	require.Contains(t, em.Message, "not found")
}
