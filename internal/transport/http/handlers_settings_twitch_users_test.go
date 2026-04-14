package httptransport

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_UpdateTwitchUser_notifyOffRejectedWhenLiveOnly(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(1)).Return(entity.TwitchUser{
		ID:                      1,
		Username:                "chan",
		IrcOnlyWhenLive:         true,
		NotifyOffStreamMessages: false,
	}, nil)

	req := &gen.UpdateTwitchUserPostRequest{ID: 1}
	req.NotifyOffStreamMessages.SetTo(true)

	res, err := h.UpdateTwitchUser(adminCtx(), req)
	require.NoError(t, err)

	bad, ok := res.(*gen.UpdateTwitchUserBadRequest)
	require.True(t, ok, "expected bad request, got %T", res)

	msg := gen.ErrorMessage(*bad).Message
	require.Contains(t, msg, "notify_off_stream_messages")
}

func TestHandler_UpdateTwitchUser_monitoredNotFoundBeforePatch(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(404)).Return(entity.TwitchUser{}, entity.ErrTwitchUserNotFound)

	req := &gen.UpdateTwitchUserPostRequest{ID: 404}
	req.Monitored.SetTo(true)

	res, err := h.UpdateTwitchUser(adminCtx(), req)
	require.NoError(t, err)

	notFound, ok := res.(*gen.UpdateTwitchUserNotFound)
	require.True(t, ok, "expected not found, got %T", res)
	require.Equal(t, "twitch user not found", notFound.Message)
}

func TestHandler_UpdateTwitchUser_monitoredSetLoadsBeforePatch(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	gomock.InOrder(
		repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(1)).Return(entity.TwitchUser{
			ID:        1,
			Username:  "chan",
			Monitored: false,
		}, nil),
		repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(1)).Return(entity.TwitchUser{
			ID:        1,
			Username:  "chan",
			Monitored: false,
		}, nil),
		repo.EXPECT().PatchTwitchUser(gomock.Any(), int64(1), gomock.Any()).Return(entity.TwitchUser{
			ID:        1,
			Username:  "chan",
			Monitored: true,
		}, nil),
	)

	req := &gen.UpdateTwitchUserPostRequest{ID: 1}
	req.Monitored.SetTo(true)

	res, err := h.UpdateTwitchUser(adminCtx(), req)
	require.NoError(t, err)

	out, ok := res.(*gen.TwitchUser)
	require.True(t, ok, "expected twitch user response, got %T", res)
	require.True(t, out.Monitored)
}
