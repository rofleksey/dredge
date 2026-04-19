package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestHandler_ListNotifications(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListNotificationEntries(gomock.Any()).Return([]entity.NotificationEntry{
		{ID: 1, Provider: "telegram", Settings: map[string]any{}, Enabled: true},
	}, nil)

	out, err := h.ListNotifications(adminCtx())
	require.NoError(t, err)
	require.Len(t, out, 1)
}
