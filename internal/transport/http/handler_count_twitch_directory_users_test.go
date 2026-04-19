package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_CountTwitchDirectoryUsers(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountTwitchUsersBrowse(gomock.Any(), gomock.Any()).Return(int64(3), nil)

	res, err := h.CountTwitchDirectoryUsers(context.Background(), gen.CountTwitchDirectoryUsersParams{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Total)
}
