package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func TestHandler_CountTwitchMessages(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountChatMessages(gomock.Any(), gomock.Any()).Return(int64(42), nil)

	res, err := h.CountTwitchMessages(context.Background(), gen.CountTwitchMessagesParams{})
	require.NoError(t, err)
	assert.Equal(t, int64(42), res.Total)
}
