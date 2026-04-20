package postgres

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// InsertIrcJoinedSample records one IRC joined count at capture time (server clock).
func (r *Repository) InsertIrcJoinedSample(ctx context.Context, joinedCount int) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.insert_irc_joined_sample")
	defer span.End()

	if joinedCount < 0 {
		joinedCount = 0
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO irc_joined_samples (joined_count) VALUES ($1)
	`, joinedCount)
	if err != nil {
		r.obs.LogError(ctx, span, "insert irc joined sample failed", err, zap.Int("joined_count", joinedCount))
	}

	return err
}

// ListIrcJoinedSamples returns samples with captured_at in [from, to], oldest first.
func (r *Repository) ListIrcJoinedSamples(ctx context.Context, from, to time.Time) ([]entity.IrcJoinedSample, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_irc_joined_samples")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, joined_count, captured_at
		FROM irc_joined_samples
		WHERE captured_at >= $1 AND captured_at <= $2
		ORDER BY captured_at ASC, id ASC
	`, from, to)
	if err != nil {
		r.obs.LogError(ctx, span, "list irc joined samples failed", err)

		return nil, err
	}
	defer rows.Close()

	out := make([]entity.IrcJoinedSample, 0)

	for rows.Next() {
		var s entity.IrcJoinedSample

		if err := rows.Scan(&s.ID, &s.JoinedCount, &s.CapturedAt); err != nil {
			r.obs.LogError(ctx, span, "scan irc joined sample failed", err)

			return nil, err
		}

		out = append(out, s)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "irc joined samples rows iteration failed", err)

		return nil, err
	}

	return out, nil
}
