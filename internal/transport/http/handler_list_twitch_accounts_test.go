package httptransport

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
)

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
