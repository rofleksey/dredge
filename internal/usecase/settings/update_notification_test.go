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

func TestService_UpdateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	en := true
	repo.EXPECT().UpdateNotificationEntry(gomock.Any(), int64(2), nil, map[string]any{"a": 1}, &en).Return(entity.NotificationEntry{ID: 2}, nil)

	_, err := svc.UpdateNotification(context.Background(), 2, nil, map[string]any{"a": 1}, &en)
	require.NoError(t, err)
}
