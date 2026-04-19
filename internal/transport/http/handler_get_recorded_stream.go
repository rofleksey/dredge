package httptransport

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetRecordedStream(ctx context.Context, params gen.GetRecordedStreamParams) (gen.GetRecordedStreamRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_recorded_stream")
	defer span.End()

	st, err := h.twitch.GetMonitoredStream(ctx, params.StreamId)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "get recorded stream failed", err)
		return nil, err
	}

	g := streamEntityToGen(st)

	return &g, nil
}
