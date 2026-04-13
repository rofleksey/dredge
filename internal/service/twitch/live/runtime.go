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

// streamLiveEdge tracks previous Helix live state for stream-start notifications.
type streamLiveEdge struct {
	initialized bool
	wasLive     bool
}

// Runtime owns the IRC monitor connection, presence polling, and notification dispatch.
type Runtime struct {
	helix                     *helix.Client
	repo                      repository.Store
	obs                       *observability.Stack
	broadcaster               interface{ BroadcastJSON(v any) }
	onEnqueue                 func(int64)
	persistParent             func() context.Context
	channelChattersSyncPeriod time.Duration
	joinReconcileInterval     time.Duration
	oauthTokenSyncInterval    time.Duration

	monitorMu     sync.Mutex
	monitorClient *twitchirc.Client

	ircMonitorMu  sync.Mutex
	ircMonitorTCP bool
	ircChannelOK  map[string]bool

	joinStateMu       sync.Mutex
	reconcilerJoined  map[string]bool
	streamEdge        map[int64]streamLiveEdge
	lastIRCOAuthToken string

	monitorLoopsMu     sync.Mutex
	monitorLoopsCancel context.CancelFunc
	monitorLoopsWG     sync.WaitGroup

	notifySem chan struct{}
}

// NewRuntime constructs runtime state for IRC-backed features.
func NewRuntime(cfg Config) *Runtime {
	period := cfg.ChannelChattersSyncPeriod
	if period <= 0 {
		period = 10 * time.Second
	}

	joinInt := cfg.JoinReconcileInterval
	if joinInt <= 0 {
		joinInt = 20 * time.Second
	}

	oauthInt := cfg.OAuthTokenSyncInterval
	if oauthInt <= 0 {
		oauthInt = 2 * time.Minute
	}

	return &Runtime{
		helix:                     cfg.Helix,
		repo:                      cfg.Repo,
		obs:                       cfg.Obs,
		broadcaster:               cfg.Broadcaster,
		onEnqueue:                 cfg.OnEnqueueUser,
		persistParent:             cfg.PersistContext,
		channelChattersSyncPeriod: period,
		joinReconcileInterval:     joinInt,
		oauthTokenSyncInterval:    oauthInt,
		reconcilerJoined:          make(map[string]bool),
		streamEdge:                make(map[int64]streamLiveEdge),
		notifySem:                 make(chan struct{}, 8),
	}
}

func (r *Runtime) persistContext() context.Context {
	if r.persistParent != nil {
		return r.persistParent()
	}
	return context.Background()
}
