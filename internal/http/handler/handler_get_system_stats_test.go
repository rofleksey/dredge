package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_GetSystemStats(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().SystemStatsTableCounts(gomock.Any()).Return(entity.SystemStatsTableCounts{
		TwitchUsers:  3,
		ChatMessages: 10,
	}, nil)

	res, err := h.GetSystemStats(adminCtx())
	require.NoError(t, err)

	body, ok := res.(*gen.SystemStatsResponse)
	require.True(t, ok)
	assert.EqualValues(t, 3, body.Tables.TwitchUsers)
	assert.EqualValues(t, 10, body.Tables.ChatMessages)
}
