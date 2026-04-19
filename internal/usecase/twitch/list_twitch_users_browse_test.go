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

func TestService_ListTwitchUsersBrowse(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	f := entity.TwitchUserBrowseFilter{Limit: 20}
	repo.EXPECT().ListTwitchUsersBrowse(gomock.Any(), f).Return([]entity.TwitchDirectoryEntry{{User: entity.TwitchUser{ID: 1}}}, nil)

	out, err := svc.ListTwitchUsersBrowse(context.Background(), f)
	require.NoError(t, err)
	assert.Len(t, out, 1)
}
