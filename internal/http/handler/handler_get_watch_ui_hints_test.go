package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetWatchUiHints_defaults(t *testing.T) {
	t.Parallel()

	h, ctrl, _ := testHandler(t)
	defer ctrl.Finish()

	res, err := h.GetWatchUiHints(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(10), res.ViewerPollIntervalSeconds)
	assert.Equal(t, int64(10), res.ChannelChattersSyncIntervalSeconds)
	assert.Equal(t, int64(60), res.MonitoredLivePollIntervalSeconds)
}
