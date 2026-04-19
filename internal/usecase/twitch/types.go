package twitch

import (
	"context"
	"sync"
	"time"

	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/gql"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
	"github.com/rofleksey/dredge/internal/service/twitch/live"
)

type Broadcaster interface {
	BroadcastJSON(v any)
}

// Usecase composes IRC chat monitoring, enrichment, notifications, and Helix API access.
type Usecase struct {
	*helix.Client
	gql  *gql.Client
	live *live.Runtime

	repo        repository.Store
	broadcaster Broadcaster
	obs         *observability.Stack
	enrichQueue chan int64

	persistMu  sync.RWMutex
	persistCtx context.Context

	viewerPollInterval          time.Duration
	channelChattersSyncInterval time.Duration
	streamSessionPollInterval   time.Duration
}

// IRCMonitorChannelStatus is one monitored channel row for the settings UI.
type IRCMonitorChannelStatus = live.IRCMonitorChannelStatus
