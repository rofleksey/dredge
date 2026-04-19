package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_ListTwitchDirectoryUsers(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListTwitchUsersBrowse(gomock.Any(), gomock.Any()).Return([]entity.TwitchDirectoryEntry{
		{User: entity.TwitchUser{ID: 1, Username: "a", Monitored: true, Marked: false}},
	}, nil)

	out, err := h.ListTwitchDirectoryUsers(context.Background(), gen.ListTwitchDirectoryUsersParams{})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "a", out[0].Username)
}
