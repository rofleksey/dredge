package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_ListNotifications(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	cur := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	curID := int64(123)
	repo.EXPECT().ListNotificationEntries(gomock.Any(), entity.NotificationListFilter{
		Limit:           20,
		CursorCreatedAt: &cur,
		CursorID:        &curID,
	}).Return([]entity.NotificationEntry{
		{ID: 1, Provider: "telegram", Settings: map[string]any{}, Enabled: true},
	}, nil)

	out, err := h.ListNotifications(adminCtx(), gen.ListNotificationsParams{
		Limit:           gen.NewOptInt(20),
		CursorCreatedAt: gen.NewOptDateTime(cur),
		CursorID:        gen.NewOptInt64(curID),
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
}
