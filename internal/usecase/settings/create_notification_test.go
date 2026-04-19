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

func TestService_CreateNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().CreateNotificationEntry(gomock.Any(), "telegram", map[string]any{}, true).Return(entity.NotificationEntry{ID: 2}, nil)

	created, err := svc.CreateNotification(context.Background(), "telegram", map[string]any{}, true)
	require.NoError(t, err)
	require.Equal(t, int64(2), created.ID)
}
