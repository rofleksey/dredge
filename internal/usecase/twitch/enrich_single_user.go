package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func (s *Usecase) enrichSingleUser(ctx context.Context, userID int64) {
	cooldown := s.enrichmentCooldown(ctx)
	if cooldown > 0 {
		_, helixFetchedAt, _, err := s.repo.GetHelixMeta(ctx, userID)
		if err == nil && helixFetchedAt != nil && time.Since(*helixFetchedAt) < cooldown {
			s.obs.Logger.Debug(
				"enrich single user: skipped by cooldown",
				zap.Int64("id", userID),
				zap.Duration("cooldown", cooldown),
			)
			return
		}
	}

	recs, err := s.HelixUsersByIDs(ctx, []int64{userID})
	if err != nil || len(recs) == 0 {
		s.obs.Logger.Debug("enrich single user: helix miss", zap.Int64("id", userID), zap.Error(err))
		return
	}

	now := time.Now().UTC()

	r := recs[0]

	var img *string
	if r.ProfileImageURL != "" {
		img = &r.ProfileImageURL
	}

	if err := s.repo.UpsertHelixMeta(ctx, r.ID, r.CreatedAt, img, now); err != nil {
		s.obs.Logger.Debug("enrich single user: upsert meta failed", zap.Int64("id", userID), zap.Error(err))
	}

	total, err := s.syncUserFollowsFromGQL(ctx, userID)
	if err != nil {
		s.obs.Logger.Debug("enrich single user: gql follows failed", zap.Int64("id", userID), zap.Error(err))
		return
	}

	if err := s.evaluateSuspicionForUser(ctx, userID, total); err != nil {
		s.obs.Logger.Debug("enrich single user: suspicion eval failed", zap.Int64("id", userID), zap.Error(err))
	}
}

func (s *Usecase) enrichmentCooldown(ctx context.Context) time.Duration {
	settings, err := s.repo.GetIrcMonitorSettings(ctx)
	if err != nil {
		return 24 * time.Hour
	}
	if settings.EnrichmentCooldown <= 0 {
		return 24 * time.Hour
	}
	return settings.EnrichmentCooldown
}
