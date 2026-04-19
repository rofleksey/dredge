package twitch

import (
	"context"
	"net/http"
	"time"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/gql"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
	"github.com/rofleksey/dredge/internal/service/twitch/live"
)

// New wires Helix, IRC runtime, and enrichment queue.
func New(repo repository.Store, broadcaster Broadcaster, cfg config.Config, obs *observability.Stack) *Usecase {
	tw := cfg.Twitch

	hx := helix.NewClient(repo, obs, tw.ClientID, tw.ClientSecret)
	hx.UserOAuthTokenCacheTTL = tw.UserOAuthTokenCacheTTL

	s := &Usecase{
		Client:                      hx,
		gql:                         gql.NewClient(http.DefaultClient),
		repo:                        repo,
		broadcaster:                 broadcaster,
		obs:                         obs,
		enrichQueue:                 make(chan int64, 10000),
		viewerPollInterval:          tw.ViewerPollInterval,
		channelChattersSyncInterval: tw.ChannelChattersSyncInterval,
		streamSessionPollInterval:   tw.StreamSessionPollInterval,
	}

	joinReconcile := 20 * time.Second

	oauthTokSync := tw.UserOAuthTokenCacheTTL / 10
	if oauthTokSync < 30*time.Second {
		oauthTokSync = 30 * time.Second
	}

	if oauthTokSync > 5*time.Minute {
		oauthTokSync = 5 * time.Minute
	}

	s.live = live.NewRuntime(live.Config{
		Helix:                     s.Client,
		Repo:                      repo,
		Obs:                       obs,
		Broadcaster:               broadcaster,
		OnEnqueueUser:             func(id int64) { s.EnqueueUserEnrichment(id) },
		PersistContext:            func() context.Context { return s.persistContext() },
		ChannelChattersSyncPeriod: s.channelChattersSyncInterval,
		JoinReconcileInterval:     joinReconcile,
		OAuthTokenSyncInterval:    oauthTokSync,
	})

	return s
}
