package twitch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_ListChatHistory_notMonitored(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "x").Return(false, nil)

	_, err := svc.ListChatHistory(context.Background(), "x", 10)
	require.ErrorIs(t, err, ErrChannelNotMonitored)
}

func TestService_ListChatHistory_ok(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "chan").Return(true, nil)
	repo.EXPECT().ListChatHistory(gomock.Any(), "chan", 5).Return([]entity.ChatHistoryMessage{{ID: 1}}, nil)

	msgs, err := svc.ListChatHistory(context.Background(), "chan", 5)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
}
