package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListRecordedStreams(ctx context.Context, params gen.ListRecordedStreamsParams) ([]gen.RecordedStream, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_recorded_streams")
	defer span.End()

	f := entity.StreamListFilter{}

	if params.Limit.IsSet() {
		f.Limit = params.Limit.Value
	}

	if params.ChannelLogin.IsSet() {
		f.ChannelLogin = params.ChannelLogin.Value
	}

	if ct, ok1 := params.CursorStartedAt.Get(); ok1 {
		if id, ok2 := params.CursorID.Get(); ok2 {
			t := ct
			f.CursorStartedAt = &t
			i := id
			f.CursorID = &i
		}
	}

	list, err := h.twitch.ListMonitoredStreams(ctx, f)
	if err != nil {
		h.obs.LogError(ctx, span, "list recorded streams failed", err)
		return nil, err
	}

	out := make([]gen.RecordedStream, 0, len(list))

	for _, s := range list {
		out = append(out, streamEntityToGen(s))
	}

	return out, nil
}
