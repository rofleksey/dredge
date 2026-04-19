package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_UpdateRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	r := entity.Rule{Regex: `bar`}
	repo.EXPECT().UpdateRule(gomock.Any(), int64(2), r).Return(entity.Rule{ID: 2}, nil)

	out, err := svc.UpdateRule(context.Background(), 2, r)
	require.NoError(t, err)
	require.Equal(t, int64(2), out.ID)
}
