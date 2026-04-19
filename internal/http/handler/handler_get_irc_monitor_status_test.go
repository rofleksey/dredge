package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetIrcMonitorStatus(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, nil)

	st, err := h.GetIrcMonitorStatus(adminCtx())
	require.NoError(t, err)
	assert.NotNil(t, st)
}

func TestHandler_GetIrcMonitorStatus_nonAdmin(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, nil)

	st, err := h.GetIrcMonitorStatus(viewerCtx())
	require.NoError(t, err)
	assert.NotNil(t, st)
}
