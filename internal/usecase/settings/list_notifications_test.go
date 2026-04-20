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

func TestService_ListNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	f := entity.NotificationListFilter{Limit: 25}
	repo.EXPECT().ListNotificationEntries(gomock.Any(), f).Return([]entity.NotificationEntry{{ID: 1}}, nil)

	list, err := svc.ListNotifications(context.Background(), f)
	require.NoError(t, err)
	require.Len(t, list, 1)
}
