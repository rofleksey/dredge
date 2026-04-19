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

func TestService_UpdateSuspicionSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	in := entity.SuspicionSettings{}
	repo.EXPECT().UpdateSuspicionSettings(gomock.Any(), in).Return(nil)
	repo.EXPECT().GetSuspicionSettings(gomock.Any()).Return(in, nil)

	out, err := svc.UpdateSuspicionSettings(context.Background(), in)
	require.NoError(t, err)
	require.Equal(t, in, out)
}
