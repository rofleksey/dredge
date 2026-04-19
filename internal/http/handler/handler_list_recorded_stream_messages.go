package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) ListRecordedStreamMessages(ctx context.Context, params gen.ListRecordedStreamMessagesParams) (gen.ListRecordedStreamMessagesRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_recorded_stream_messages")
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

	f := entity.ChatMessageListFilter{Limit: limit}

	if params.Username.IsSet() {
		f.Username = params.Username.Value
	}

	if params.Text.IsSet() {
		f.Text = params.Text.Value
	}

	if params.ChatterUserID.IsSet() {
		v := params.ChatterUserID.Value
		f.ChatterUserID = &v
	}

	if ct, ok1 := params.CursorCreatedAt.Get(); ok1 {
		if id, ok2 := params.CursorID.Get(); ok2 {
			t := ct
			f.CursorCreatedAt = &t
			i := id
			f.CursorID = &i
		}
	}

	list, err := h.twitch.ListStreamMessages(ctx, params.StreamId, f)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "list recorded stream messages failed", err)
		return nil, err
	}

	out := make(gen.ListRecordedStreamMessagesOKApplicationJSON, 0, len(list))

	for _, m := range list {
		out = append(out, chatHistoryEntityToGen(m))
	}

	return &out, nil
}
