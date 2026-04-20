package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_ListIrcMonitorJoinedHistory(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	ts := time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC)

	repo.EXPECT().ListIrcJoinedSamples(gomock.Any(), gomock.Any(), gomock.Any()).Return([]entity.IrcJoinedSample{
		{ID: 1, CapturedAt: ts, JoinedCount: 27},
	}, nil)

	params := gen.ListIrcMonitorJoinedHistoryParams{}
	params.Days.SetTo(7)

	out, err := h.ListIrcMonitorJoinedHistory(adminCtx(), params)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, 27, out[0].JoinedCount)
	assert.True(t, out[0].CapturedAt.Equal(ts))
}

func TestHandler_ListIrcMonitorJoinedHistory_nonAdmin(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListIrcJoinedSamples(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	params := gen.ListIrcMonitorJoinedHistoryParams{}
	params.Days.SetTo(7)

	out, err := h.ListIrcMonitorJoinedHistory(viewerCtx(), params)
	require.NoError(t, err)
	assert.Empty(t, out)
}
