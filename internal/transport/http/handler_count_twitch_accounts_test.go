package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandler_CountTwitchAccounts(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountTwitchAccounts(gomock.Any()).Return(int64(2), nil)

	res, err := h.CountTwitchAccounts(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
}
