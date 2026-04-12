package live

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

// Config wires the anonymous IRC monitor, presence snapshots, notifications, and outbound send.
type Config struct {
	Helix                     *helix.Client
	Repo                      repository.Store
	Obs                       *observability.Stack
	Broadcaster               interface{ BroadcastJSON(v any) }
	OnEnqueueUser             func(userID int64)
	PersistContext            func() context.Context
	ChannelChattersSyncPeriod time.Duration
}
