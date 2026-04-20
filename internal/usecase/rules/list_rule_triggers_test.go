package rules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestUsecase_ListRuleTriggers(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := NewUsecase(repo, obs, nil, nil)

	ts := time.Unix(100, 0).UTC()
	repo.EXPECT().ListRuleTriggerEvents(gomock.Any(), entity.RuleTriggerListFilter{Limit: 5}).Return([]entity.RuleTriggerEvent{
		{ID: 1, CreatedAt: ts, RuleName: "r", TriggerEvent: "chat_message", ActionType: "notify", DisplayText: "x"},
	}, nil)

	out, err := svc.ListRuleTriggers(context.Background(), entity.RuleTriggerListFilter{Limit: 5})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "x", out[0].DisplayText)
}
