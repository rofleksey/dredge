package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const maxFollowerPages = 25

// RunHelixEnrichment refreshes account creation times and follow relationships (best-effort).
func (s *Service) RunHelixEnrichment(ctx context.Context) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.helix_enrichment")
	defer span.End()

	s.obs.Logger.Info("starting helix enrichment job")

	ids, err := s.repo.ListDistinctChattersWithMessages(ctx, 800)
	if err != nil {
		s.obs.LogError(ctx, span, "list chatters for enrichment failed", err)
		return
	}

	if len(ids) > 0 {
		recs, err := s.HelixUsersByIDs(ctx, ids)
		if err != nil {
			s.obs.LogError(ctx, span, "helix users by id batch failed", err)
		} else {
			now := time.Now().UTC()

			for _, r := range recs {
				var img *string
				if r.ProfileImageURL != "" {
					img = &r.ProfileImageURL
				}

				if err := s.repo.UpsertHelixMeta(ctx, r.ID, r.CreatedAt, img, now); err != nil {
					s.obs.Logger.Warn("upsert helix meta failed", zap.Int64("id", r.ID), zap.Error(err))
				}
			}
		}
	}

	accessToken, acc, err := s.accessTokenForFirstLinkedAccount(ctx)
	if err != nil {
		s.obs.Logger.Debug("helix enrichment: no twitch accounts for follower lookup", zap.Error(err))
		return
	}

	modID, err := s.ResolveUserIDByLogin(ctx, acc.Username)
	if err != nil {
		s.obs.LogError(ctx, span, "resolve moderator id for enrichment failed", err)
		return
	}

	pairs, err := s.repo.ListChatterChannelPairsForFollowEnrichment(ctx, 800)
	if err != nil {
		s.obs.LogError(ctx, span, "list chatter/channel pairs failed", err)
		return
	}

	now := time.Now().UTC()

	for _, p := range pairs {
		fa := s.findFollowedAtForChatter(ctx, accessToken, p.ChannelID, modID, p.ChatterID)
		if err := s.repo.UpsertChannelFollow(ctx, p.ChatterID, p.ChannelID, fa, now); err != nil {
			s.obs.Logger.Warn("upsert channel follow failed", zap.Error(err))
		}
	}

	for _, uid := range ids {
		total, err := s.syncUserFollowsFromGQL(ctx, uid)
		if err != nil {
			s.obs.Logger.Debug("gql follows sync skipped", zap.Int64("user_id", uid), zap.Error(err))
			continue
		}

		if err := s.evaluateSuspicionForUser(ctx, uid, total); err != nil {
			s.obs.Logger.Debug("suspicion evaluation failed", zap.Int64("user_id", uid), zap.Error(err))
		}
	}

	s.obs.Logger.Info("helix enrichment job finished")
}

func (s *Service) findFollowedAtForChatter(ctx context.Context, accessToken string, broadcasterID, moderatorID, chatterID int64) *time.Time {
	var cursor string

	for page := 0; page < maxFollowerPages; page++ {
		rows, next, err := s.FetchChannelFollowersPage(ctx, accessToken, broadcasterID, moderatorID, cursor)
		if err != nil {
			s.obs.Logger.Debug("followers page failed",
				zap.Int64("broadcaster_id", broadcasterID),
				zap.Error(err))
			return nil
		}

		for _, row := range rows {
			if row.UserID == chatterID {
				return row.FollowedAt
			}
		}

		if next == "" {
			break
		}

		cursor = next
	}

	return nil
}
