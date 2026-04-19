package settings

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
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_UpdateIrcMonitorSettings_rejectsUnknownAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	id := int64(99)
	repo.EXPECT().GetTwitchAccountByID(gomock.Any(), id).Return(entity.TwitchAccount{}, entity.ErrTwitchAccountNotFound)

	_, err := svc.UpdateIrcMonitorSettings(context.Background(), entity.IrcMonitorSettings{
		OauthTwitchAccountID: &id,
		EnrichmentCooldown:   24 * time.Hour,
	})
	require.ErrorIs(t, err, entity.ErrTwitchAccountNotFound)
}

func TestService_UpdateIrcMonitorSettings_defaultsCooldown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	in := entity.IrcMonitorSettings{OauthTwitchAccountID: nil}
	expected := entity.IrcMonitorSettings{
		OauthTwitchAccountID: nil,
		EnrichmentCooldown:   24 * time.Hour,
	}

	repo.EXPECT().UpdateIrcMonitorSettings(gomock.Any(), expected).Return(nil)
	repo.EXPECT().GetIrcMonitorSettings(gomock.Any()).Return(expected, nil)

	out, err := svc.UpdateIrcMonitorSettings(context.Background(), in)
	require.NoError(t, err)
	require.Equal(t, expected.EnrichmentCooldown, out.EnrichmentCooldown)
}
