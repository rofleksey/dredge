package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListTwitchUsersBrowse lists known Twitch identities for the directory UI.
func (s *Usecase) ListTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) ([]entity.TwitchDirectoryEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_twitch_users_browse")
	defer span.End()

	list, err := s.repo.ListTwitchUsersBrowse(ctx, f)
	if err != nil {
		s.obs.LogError(ctx, span, "list twitch users browse failed", err)
		return nil, err
	}

	return list, nil
}
