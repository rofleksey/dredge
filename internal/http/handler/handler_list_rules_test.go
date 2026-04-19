package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/authctx"
)

func TestHandler_ListRules_admin(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListRules(gomock.Any()).Return([]entity.Rule{{ID: 1, Regex: "x"}}, nil)

	ctx := authctx.WithRole(authctx.WithUserID(context.Background(), int64(1)), "admin")

	out, err := h.ListRules(ctx)
	require.NoError(t, err)
	require.Len(t, out, 1)
}
