package twitch

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestRunHelixEnrichment_noChatters_stopsEarly(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	repo.EXPECT().ListDistinctChattersWithMessages(gomock.Any(), 800).Return(nil, nil)
	repo.EXPECT().ListTwitchAccounts(gomock.Any()).Return(nil, nil)

	svc.RunHelixEnrichment(context.Background())
}

func TestRunHelixEnrichment_withChatters_noAccounts(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	repo.EXPECT().ListDistinctChattersWithMessages(gomock.Any(), 800).Return([]int64{1}, nil)
	repo.EXPECT().ListTwitchAccounts(gomock.Any()).Return(nil, nil)

	svc.RunHelixEnrichment(context.Background())
}

func TestRunHelixEnrichment_listChattersError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	repo.EXPECT().ListDistinctChattersWithMessages(gomock.Any(), 800).Return(nil, assertErr{})

	svc.RunHelixEnrichment(context.Background())
}

type assertErr struct{}

func (assertErr) Error() string { return "assert" }

func TestRunHelixEnrichment_accountsButRefreshFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	repo.EXPECT().ListDistinctChattersWithMessages(gomock.Any(), 800).Return(nil, nil)
	repo.EXPECT().ListTwitchAccounts(gomock.Any()).Return([]entity.TwitchAccount{
		{ID: 1, Username: "u", RefreshToken: "bad"},
	}, nil)

	svc.RunHelixEnrichment(context.Background())
}
