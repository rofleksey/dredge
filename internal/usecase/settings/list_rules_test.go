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

func TestService_ListRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().ListRules(gomock.Any()).Return([]entity.Rule{{ID: 1, Regex: "x"}}, nil)

	rules, err := svc.ListRules(context.Background())
	require.NoError(t, err)
	require.Len(t, rules, 1)
}
