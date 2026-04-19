package twitch

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

// GetMonitoredStream returns a stream row if the channel is monitored.
func (s *Service) GetMonitoredStream(ctx context.Context, id int64) (entity.Stream, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.get_monitored_stream")
	defer span.End()

	st, err := s.repo.GetMonitoredStreamByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Stream{}, entity.ErrStreamNotFound
	}

	return st, err
}
