package twitch

import (
	"context"
	"net/http"
	"time"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/service/twitch/gql"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
	"github.com/rofleksey/dredge/internal/service/twitch/live"
	"go.uber.org/zap"
)

// New wires Helix, IRC runtime, and enrichment queue.
func New(repo repository.Store, broadcaster Broadcaster, cfg config.Config, obs *observability.Stack) *Service {
	tw := cfg.Twitch

	hx := helix.NewClient(repo, obs, tw.ClientID, tw.ClientSecret)
	hx.UserOAuthTokenCacheTTL = tw.UserOAuthTokenCacheTTL

	s := &Service{
		Client:                      hx,
		gql:                         gql.NewClient(http.DefaultClient),
		repo:                        repo,
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

// WatchUiHints exposes poll intervals for the SPA (seconds, minimum 1).
func (s *Service) WatchUiHints() (viewerPollSec int, channelChattersSyncSec int) {
	v := int(s.viewerPollInterval / time.Second)
	c := int(s.channelChattersSyncInterval / time.Second)

	if v < 1 {
		v = 10
	}

	if c < 1 {
		c = 10
	}

	return v, c
}

// SetPersistContext sets the parent context for IRC-driven DB work; cancel it on app shutdown.
func (s *Service) SetPersistContext(ctx context.Context) {
	s.persistMu.Lock()
	defer s.persistMu.Unlock()

	s.persistCtx = ctx
}

func (s *Service) persistContext() context.Context {
	s.persistMu.RLock()
	defer s.persistMu.RUnlock()

	if s.persistCtx != nil {
		return s.persistCtx
	}

	return context.Background()
}

func (s *Service) StartMonitor(ctx context.Context) error {
	return s.live.StartMonitor(ctx)
}

func (s *Service) RestartMonitor(ctx context.Context) error {
	return s.live.RestartMonitor(ctx)
}

func (s *Service) StopMonitor() {
	s.live.StopMonitor()
}

func (s *Service) SendMessage(ctx context.Context, accountID int64, channel, message string) error {
	return s.live.SendMessage(ctx, accountID, channel, message)
}

func (s *Service) StartPresenceTicker(ctx context.Context) {
	s.live.StartPresenceTicker(ctx)
}

func (s *Service) GetIrcMonitorStatus(ctx context.Context) (connected bool, channels []IRCMonitorChannelStatus, err error) {
	return s.live.GetIrcMonitorStatus(ctx)
}

// LiveWebSocketWelcomePayloads returns messages sent to a browser immediately after the live WebSocket upgrade.
func (s *Service) LiveWebSocketWelcomePayloads(ctx context.Context) ([]any, error) {
	msg, err := s.live.LiveWebSocketWelcomePayloads(ctx)
	if err != nil {
		return nil, err
	}

	return []any{msg}, nil
}

// GetTwitchUser returns a twitch_users row by id.
func (s *Service) GetTwitchUser(ctx context.Context, id int64) (entity.TwitchUser, error) {
	return s.repo.GetTwitchUserByID(ctx, id)
}

// ListChatHistory returns persisted messages for a monitored channel (oldest first).
func (s *Service) ListChatHistory(ctx context.Context, channel string, limit int) ([]entity.ChatHistoryMessage, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_chat_history")
	defer span.End()

	ok, err := s.repo.IsMonitoredChannel(ctx, channel)
	if err != nil {
		s.obs.LogError(ctx, span, "check monitored channel failed", err)
		return nil, err
	}

	if !ok {
		return nil, ErrChannelNotMonitored
	}

	list, err := s.repo.ListChatHistory(ctx, channel, limit)
	if err != nil {
		s.obs.LogError(ctx, span, "list chat history failed", err)
		return nil, err
	}

	return list, nil
}

// ListChatMessages returns persisted messages matching filters (newest first).
func (s *Service) ListChatMessages(ctx context.Context, f entity.ChatMessageListFilter) ([]entity.ChatHistoryMessage, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_chat_messages")
	defer span.End()

	list, err := s.repo.ListChatMessages(ctx, f)
	if err != nil {
		s.obs.LogError(ctx, span, "list chat messages failed", err)
		return nil, err
	}

	return list, nil
}

// ListTwitchUsersBrowse lists known Twitch identities for the directory UI.
func (s *Service) ListTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) ([]entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_twitch_users_browse")
	defer span.End()

	list, err := s.repo.ListTwitchUsersBrowse(ctx, f)
	if err != nil {
		s.obs.LogError(ctx, span, "list twitch users browse failed", err)
		return nil, err
	}

	return list, nil
}

func startOfWeekMondayUTC(t time.Time) time.Time {
	t = t.UTC()
	wd := int(t.Weekday())
	daysSinceMon := (wd + 6) % 7
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return start.AddDate(0, 0, -daysSinceMon)
}

func (s *Service) presenceSecondsThisWeek(ctx context.Context, chatterID int64) (int64, error) {
	now := time.Now().UTC()
	from := startOfWeekMondayUTC(now)

	evs, err := s.repo.ListUserActivityEventsForTimeline(ctx, chatterID, from, now)
	if err != nil {
		return 0, err
	}

	segs := BuildActivityTimelineSegments(evs, now)

	var sum int64

	for _, seg := range segs {
		if seg.End.After(seg.Start) {
			sum += int64(seg.End.Sub(seg.Start).Seconds())
		}
	}

	return sum, nil
}

// GetTwitchUserProfile returns profile fields, message count, IRC presence seconds this week (UTC Mon..now), helix created_at, profile image, monitored follows, GQL full follows list, and global blacklist for UI.
func (s *Service) GetTwitchUserProfile(ctx context.Context, id int64) (
	u entity.TwitchUser,
	messageCount int64,
	presenceSec int64,
	accountCreated *time.Time,
	profileImageURL *string,
	monitoredFollows []entity.FollowedMonitoredChannel,
	gqlFollows []entity.FollowedChannelRow,
	channelBlacklist []string,
	err error,
) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.get_twitch_user_profile")
	defer span.End()

	u, err = s.repo.GetTwitchUserByID(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "get twitch user failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	messageCount, err = s.repo.CountChatMessagesByChatter(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "count chatter messages failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	presenceSec, err = s.presenceSecondsThisWeek(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "presence this week failed", err, zap.Int64("id", id))

		presenceSec = 0
	}

	accountCreated, _, profileImageURL, err = s.repo.GetHelixMeta(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "get helix meta failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	monitoredFollows, err = s.repo.ListFollowedMonitoredChannels(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "list follows failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	gqlFollows, err = s.repo.ListUserFollowedChannels(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "list gql follows failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	channelBlacklist, err = s.repo.ListChannelBlacklist(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list blacklist failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	return u, messageCount, presenceSec, accountCreated, profileImageURL, monitoredFollows, gqlFollows, channelBlacklist, nil
}

// CountChatMessages delegates to repository (same filters as list, no cursor).
func (s *Service) CountChatMessages(ctx context.Context, f entity.ChatMessageListFilter) (int64, error) {
	return s.repo.CountChatMessages(ctx, f)
}

// CountTwitchUsersBrowse delegates to repository.
func (s *Service) CountTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) (int64, error) {
	return s.repo.CountTwitchUsersBrowse(ctx, f)
}
