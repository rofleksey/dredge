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

func TestService_GetIrcMonitorSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	want := entity.IrcMonitorSettings{}
	repo.EXPECT().GetIrcMonitorSettings(gomock.Any()).Return(want, nil)

	out, err := svc.GetIrcMonitorSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, want, out)
}
