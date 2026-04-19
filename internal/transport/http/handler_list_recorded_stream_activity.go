package httptransport

import (
	"context"
	"errors"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListRecordedStreamActivity(ctx context.Context, params gen.ListRecordedStreamActivityParams) (gen.ListRecordedStreamActivityRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_recorded_stream_activity")
	defer span.End()

	limit := 50
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 200 {
		limit = 200
	}

	var (
		cursorCreatedAt *time.Time
		cursorID        *int64
	)

	if ct, ok1 := params.CursorCreatedAt.Get(); ok1 {
		if id, ok2 := params.CursorID.Get(); ok2 {
			t := ct
			cursorCreatedAt = &t
			i := id
			cursorID = &i
		}
	}

	list, err := h.twitch.ListStreamActivity(ctx, params.StreamId, limit, cursorCreatedAt, cursorID)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "list recorded stream activity failed", err)
		return nil, err
	}

	out := make(gen.ListRecordedStreamActivityOKApplicationJSON, 0, len(list))

	for _, e := range list {
		out = append(out, entityActivityToGen(e, ""))
	}

	return &out, nil
}
