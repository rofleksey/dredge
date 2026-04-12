package live

import (
	"context"
	"sync"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

// Runtime owns the anonymous IRC monitor connection, presence polling, and notification dispatch.
type Runtime struct {
	helix                     *helix.Client
	repo                      repository.Store
	obs                       *observability.Stack
	broadcaster               interface{ BroadcastJSON(v any) }
	onEnqueue                 func(int64)
	persistParent             func() context.Context
	channelChattersSyncPeriod time.Duration

	monitorMu     sync.Mutex
	monitorClient *twitchirc.Client

	ircMonitorMu  sync.Mutex
	ircMonitorTCP bool
	ircChannelOK  map[string]bool

	notifySem chan struct{}
}

// NewRuntime constructs runtime state for IRC-backed features.
func NewRuntime(cfg Config) *Runtime {
	period := cfg.ChannelChattersSyncPeriod
	if period <= 0 {
		period = 10 * time.Second
	}

	return &Runtime{
		helix:                     cfg.Helix,
		repo:                      cfg.Repo,
		obs:                       cfg.Obs,
		broadcaster:               cfg.Broadcaster,
		onEnqueue:                 cfg.OnEnqueueUser,
		persistParent:             cfg.PersistContext,
		channelChattersSyncPeriod: period,
		notifySem:                 make(chan struct{}, 8),
	}
}

func (r *Runtime) persistContext() context.Context {
	if r.persistParent != nil {
		return r.persistParent()
	}
	return context.Background()
}
