package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func (s *Usecase) syncStreamSessions(ctx context.Context) error {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.sync_stream_sessions")
	defer span.End()

	monitored, err := s.repo.ListMonitoredTwitchUsers(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list monitored for stream recorder failed", err)
		return err
	}

	ids := make([]int64, 0, len(monitored))
	for _, u := range monitored {
		ids = append(ids, u.ID)
	}

	live, err := s.HelixStreamsMetadataByBroadcasterIDs(ctx, ids)
	if err != nil {
		s.obs.LogError(ctx, span, "helix streams metadata failed", err)
		return err
	}

	for _, snap := range live {
		if snap.HelixStreamID == "" {
			continue
		}

		st := snap.StartedAt
		if st.IsZero() {
			st = time.Now().UTC()
		}

		vc := snap.ViewerCount
		if _, err := s.repo.UpsertStreamFromHelix(ctx, snap.UserID, snap.HelixStreamID, st, snap.Title, snap.GameName, &vc); err != nil {
			s.obs.Logger.Warn("upsert stream from helix failed",
				zap.Error(err), zap.Int64("channel_user_id", snap.UserID), zap.String("helix_stream_id", snap.HelixStreamID))
		}
	}

	for _, u := range monitored {
		if _, ok := live[u.ID]; ok {
			continue
		}

		if err := s.repo.CloseOpenStreamsForChannel(ctx, u.ID); err != nil {
			s.obs.Logger.Warn("close open streams for offline channel failed", zap.Error(err), zap.Int64("channel_user_id", u.ID))
		}
	}

	return nil
}
