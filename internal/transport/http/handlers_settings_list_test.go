package httptransport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_ListTwitchUsers_settings(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListTwitchUsers(gomock.Any()).Return([]entity.TwitchUser{
		{ID: 1, Username: "a", Monitored: true},
	}, nil)

	out, err := h.ListTwitchUsers(adminCtx(), gen.ListTwitchUsersParams{})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "a", out[0].Username)
}

func TestHandler_ListTwitchUsers_settings_monitoredOnly(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return([]entity.TwitchUser{
		{ID: 2, Username: "b", Monitored: true},
	}, nil)

	var p gen.ListTwitchUsersParams
	p.MonitoredOnly.SetTo(true)

	out, err := h.ListTwitchUsers(adminCtx(), p)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "b", out[0].Username)
}

func TestHandler_ListTwitchAccounts_settings(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListTwitchAccounts(gomock.Any()).Return([]entity.TwitchAccount{
		{ID: 1, Username: "bot", AccountType: "bot"},
	}, nil)

	out, err := h.ListTwitchAccounts(adminCtx())
	require.NoError(t, err)
	require.Len(t, out, 1)
}
