package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// GetTwitchUser returns a twitch_users row by id.
func (s *Usecase) GetTwitchUser(ctx context.Context, id int64) (entity.TwitchUser, error) {
	return s.repo.GetTwitchUserByID(ctx, id)
}
