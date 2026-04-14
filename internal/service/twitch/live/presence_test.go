package live

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestIsGoTwitchIRCUserlistMissing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "other", err: errors.New("network down"), want: false},
		{
			name: "library_message",
			err:  fmt.Errorf("Could not find userlist for channel 'foo' in client"),
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := isGoTwitchIRCUserlistMissing(tc.err); got != tc.want {
				t.Fatalf("isGoTwitchIRCUserlistMissing(...) = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSnapshotChannelPresence_offlineClearsAndEmitsLeaveEvents(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	r := NewRuntime(Config{
		Repo: repo,
		Obs:  obs,
	})

	channel := entity.TwitchUser{ID: 77, Username: "offline_channel"}

	repo.EXPECT().ListChannelChatterIDs(gomock.Any(), int64(77)).Return([]int64{101, 102}, nil)
	repo.EXPECT().InsertUserActivityEvent(gomock.Any(), int64(101), entity.UserActivityChatOffline, gomock.Any(), nil).Return(nil)
	repo.EXPECT().InsertUserActivityEvent(gomock.Any(), int64(102), entity.UserActivityChatOffline, gomock.Any(), nil).Return(nil)
	repo.EXPECT().ReplaceChannelChattersSnapshot(gomock.Any(), int64(77), gomock.Len(0)).Return(nil)

	err := r.snapshotChannelPresence(context.Background(), nil, channel, "offline_channel", false)
	if err != nil {
		t.Fatalf("snapshotChannelPresence() error = %v", err)
	}
}

func TestEmitPresenceDiffEvents_emitsJoinAndLeaveEdges(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	r := NewRuntime(Config{
		Repo: repo,
		Obs:  obs,
	})

	prevSet := map[int64]struct{}{
		100: {},
		101: {},
	}
	currSet := map[int64]struct{}{
		101: {},
		102: {},
	}

	repo.EXPECT().InsertUserActivityEvent(gomock.Any(), int64(102), entity.UserActivityChatOnline, gomock.Any(), nil).Return(nil)
	repo.EXPECT().InsertUserActivityEvent(gomock.Any(), int64(100), entity.UserActivityChatOffline, gomock.Any(), nil).Return(nil)

	r.emitPresenceDiffEvents(context.Background(), 77, prevSet, currSet)
}
