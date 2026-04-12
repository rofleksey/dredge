package twitch

import (
	"context"
	"strings"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

func isAutoSusType(t *string) bool {
	if t == nil {
		return false
	}

	switch *t {
	case entity.SusTypeAutoAge, entity.SusTypeAutoBlacklist, entity.SusTypeAutoLowFollow:
		return true
	default:
		return false
	}
}

// evaluateSuspicionForUser applies automatic suspicion rules after enrichment data is fresh.
// gqlTotalCount is total follows from Twitch (may exceed stored rows if pagination capped).
func (s *Service) evaluateSuspicionForUser(ctx context.Context, userID int64, gqlTotalCount int) error {
	ownIDs, err := s.repo.ListLinkedTwitchAccountUserIDs(ctx)
	if err != nil {
		return err
	}

	for _, id := range ownIDs {
		if id == userID {
			return s.clearOwnAccountSuspicion(ctx, userID)
		}
	}

	u, err := s.repo.GetTwitchUserByID(ctx, userID)
	if err != nil {
		return err
	}

	settings, err := s.repo.GetSuspicionSettings(ctx)
	if err != nil {
		return err
	}

	blacklist, err := s.repo.ListChannelBlacklist(ctx)
	if err != nil {
		return err
	}

	blSet := make(map[string]struct{}, len(blacklist))
	for _, login := range blacklist {
		blSet[strings.ToLower(login)] = struct{}{}
	}

	follows, err := s.repo.ListUserFollowedChannels(ctx, userID)
	if err != nil {
		return err
	}

	accountCreated, _, err := s.repo.GetHelixMeta(ctx, userID)
	if err != nil {
		return err
	}

	shouldSus, typ, desc := computeAutoSuspicion(settings, blSet, follows, gqlTotalCount, accountCreated, time.Now().UTC())

	manualLocked := u.SusType != nil && *u.SusType == entity.SusTypeManual && u.IsSus

	// User dismissed auto suspicion: never auto-mark true again until they re-enable.
	if u.SusAutoSuppressed {
		if shouldSus {
			return nil
		}
		// Predicates no longer match: clear auto-tagged suspicion only.
		if u.IsSus && isAutoSusType(u.SusType) {
			return s.applySuspicionPatch(ctx, userID, false, nil, nil)
		}
		return nil
	}

	if manualLocked {
		// Manual suspicious: do not auto-clear based on predicates.
		return nil
	}

	if shouldSus {
		st := typ
		sd := desc
		return s.applySuspicionPatch(ctx, userID, true, &st, &sd)
	}

	// Not suspicious: clear only auto-derived marks.
	if u.IsSus && isAutoSusType(u.SusType) {
		return s.applySuspicionPatch(ctx, userID, false, nil, nil)
	}

	return nil
}

func (s *Service) clearOwnAccountSuspicion(ctx context.Context, userID int64) error {
	u, err := s.repo.GetTwitchUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !u.IsSus && u.SusType == nil {
		return nil
	}
	return s.applySuspicionPatch(ctx, userID, false, nil, nil)
}

func (s *Service) applySuspicionPatch(ctx context.Context, userID int64, isSus bool, susType *string, susDesc *string) error {
	empty := ""

	p := entity.TwitchUserPatch{IsSus: &isSus}
	if isSus {
		p.SusType = susType
		p.SusDescription = susDesc
	} else {
		p.SusType = &empty
		p.SusDescription = &empty
	}

	_, err := s.repo.PatchTwitchUser(ctx, userID, p)

	return err
}

// computeAutoSuspicion returns whether the user matches automatic suspicion (priority: blacklist > account age > low follow count).
func computeAutoSuspicion(
	settings entity.SuspicionSettings,
	blacklist map[string]struct{},
	follows []entity.FollowedChannelRow,
	gqlTotalCount int,
	accountCreated *time.Time,
	now time.Time,
) (should bool, susType string, description string) {
	if settings.AutoCheckBlacklist {
		for _, f := range follows {
			if _, ok := blacklist[strings.ToLower(f.FollowedChannelLogin)]; ok {
				return true, entity.SusTypeAutoBlacklist, "Follows a channel on the blacklist"
			}
		}
	}

	if settings.AutoCheckAccountAge && accountCreated != nil {
		threshold := time.Duration(settings.AccountAgeSusDays) * 24 * time.Hour
		if settings.AccountAgeSusDays > 0 && now.Sub(*accountCreated) < threshold {
			return true, entity.SusTypeAutoAge, "Account was created within the configured recent window"
		}
	}

	if settings.AutoCheckLowFollows {
		th := settings.LowFollowsThreshold
		if th < 0 {
			th = 0
		}

		if gqlTotalCount < th {
			return true, entity.SusTypeAutoLowFollow, "Follow count is below the configured minimum"
		}
	}

	return false, "", ""
}
