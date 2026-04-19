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

func TestCreateRuleRejectsInvalidRegex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)

	_, err := svc.CreateRule(context.Background(), entity.Rule{Regex: "("})
	require.Error(t, err)
}

func TestService_CreateRule_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	r := entity.Rule{Regex: `foo`}
	repo.EXPECT().CreateRule(gomock.Any(), r).Return(entity.Rule{ID: 1, Regex: "foo"}, nil)

	out, err := svc.CreateRule(context.Background(), r)
	require.NoError(t, err)
	require.Equal(t, int64(1), out.ID)
}
