package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// CountTwitchUsersBrowse delegates to repository.
func (s *Service) CountTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) (int64, error) {
	return s.repo.CountTwitchUsersBrowse(ctx, f)
}
