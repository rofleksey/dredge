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

func TestService_CreateTwitchAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().CreateTwitchAccount(gomock.Any(), int64(2), "u", "rt", "bot").Return(entity.TwitchAccount{ID: 2}, nil)

	a, err := svc.CreateTwitchAccount(context.Background(), 2, "u", "rt", "bot")
	require.NoError(t, err)
	require.Equal(t, int64(2), a.ID)
}
