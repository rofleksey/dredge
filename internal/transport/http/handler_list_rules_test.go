package httptransport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestHandler_ListRules_admin(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListRules(gomock.Any()).Return([]entity.Rule{{ID: 1, Regex: "x"}}, nil)

	ctx := context.WithValue(context.WithValue(context.Background(), userIDCtxKey, int64(1)), roleCtxKey, "admin")

	out, err := h.ListRules(ctx)
	require.NoError(t, err)
	require.Len(t, out, 1)
}
