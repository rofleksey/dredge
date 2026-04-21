package twitch

import (
	"context"
	"strings"

	"go.uber.org/zap"
)

// RunChannelDiscovery polls Helix live streams for the configured game and upserts matching non-monitored candidates.
func (s *Usecase) RunChannelDiscovery(ctx context.Context) error {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.run_channel_discovery")
	defer span.End()

	st, err := s.repo.GetChannelDiscoverySettings(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "channel discovery settings load failed", err)
		return err
	}

	if !st.Enabled {
		return nil
	}

	gameID := strings.TrimSpace(st.GameID)
	if gameID == "" {
		s.obs.Logger.Warn("channel discovery enabled but game_id is empty; skipping run")
		return nil
	}

	deniedIDs, err := s.repo.ListTwitchDiscoveryDeniedUserIDs(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list discovery denied failed", err)
		return err
	}

	denied := make(map[int64]struct{}, len(deniedIDs))
	for _, id := range deniedIDs {
		denied[id] = struct{}{}
	}

	monitored, err := s.repo.ListMonitoredTwitchUsers(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list monitored for discovery failed", err)
		return err
	}

	monitoredSet := make(map[int64]struct{}, len(monitored))
	for _, u := range monitored {
		monitoredSet[u.ID] = struct{}{}
	}

	required := st.RequiredStreamTags
	minV := int64(st.MinLiveViewers)

	maxPages := st.MaxStreamPagesPerRun

	if maxPages < 1 {
		maxPages = 1
	}

	cursor := ""

	for page := 0; page < maxPages; page++ {
		rows, next, err := s.HelixStreamsByGameIDPage(ctx, gameID, 100, cursor)
		if err != nil {
			s.obs.LogError(ctx, span, "helix streams by game failed", err, zap.Int("page", page))
			return err
		}

		for _, row := range rows {
			if _, ok := denied[row.UserID]; ok {
				continue
			}

			if _, ok := monitoredSet[row.UserID]; ok {
				continue
			}

			if row.ViewerCount < minV {
				continue
			}

			if !streamTagsCoverRequired(required, row.Tags) {
				continue
			}

			if _, err := s.repo.UpsertTwitchUserFromChat(ctx, row.UserID, row.UserLogin); err != nil {
				s.obs.Logger.Warn("discovery upsert twitch user failed",
					zap.Error(err), zap.Int64("twitch_user_id", row.UserID), zap.String("login", row.UserLogin))
				continue
			}

			var titlePtr, gamePtr *string

			if t := strings.TrimSpace(row.Title); t != "" {
				titlePtr = &row.Title
			}

			if g := strings.TrimSpace(row.GameName); g != "" {
				gamePtr = &row.GameName
			}

			vc := row.ViewerCount

			if err := s.repo.UpsertTwitchDiscoveryCandidate(ctx, row.UserID, &vc, titlePtr, gamePtr, row.Tags); err != nil {
				s.obs.Logger.Warn("discovery upsert candidate failed",
					zap.Error(err), zap.Int64("twitch_user_id", row.UserID))
			}
		}

		if strings.TrimSpace(next) == "" {
			break
		}

		cursor = next
	}

	return nil
}
