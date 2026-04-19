package rules

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

func TestUsecase_Engine(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(nil, obs, nil, nil)
	require.Nil(t, uc.Engine())
}

func TestUsecase_Bootstrap_nilEngine(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(nil, obs, nil, nil)
	require.NoError(t, uc.Bootstrap(context.Background()))
}

func TestUsecase_ListRules(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(repo, obs, nil, nil)

	want := []entity.Rule{{ID: 1, Name: "n", EventType: EventChatMessage}}
	repo.EXPECT().ListRules(gomock.Any()).Return(want, nil)

	out, err := uc.ListRules(context.Background())
	require.NoError(t, err)
	require.Equal(t, want, out)
}

func TestUsecase_CountRules(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(repo, obs, nil, nil)

	repo.EXPECT().CountRules(gomock.Any()).Return(int64(3), nil)

	n, err := uc.CountRules(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(3), n)
}

func TestUsecase_CreateRule_validationError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(repo, obs, nil, nil)

	_, err := uc.CreateRule(context.Background(), entity.Rule{})
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}

func TestUsecase_UpdateRule_validationError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(repo, obs, nil, nil)

	_, err := uc.UpdateRule(context.Background(), 1, entity.Rule{})
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}

func TestUsecase_DeleteRule_notFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	uc := NewUsecase(repo, obs, nil, nil)

	repo.EXPECT().DeleteRule(gomock.Any(), int64(99)).Return(entity.ErrRuleNotFound)

	err := uc.DeleteRule(context.Background(), 99)
	require.ErrorIs(t, err, entity.ErrRuleNotFound)
}
