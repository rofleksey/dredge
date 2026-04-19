package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/http/authctx"
)

func TestHandler_CountRules_admin(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountRules(gomock.Any()).Return(int64(5), nil)

	ctx := authctx.WithRole(authctx.WithUserID(context.Background(), int64(1)), "admin")

	res, err := h.CountRules(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(5), res.Total)
}
