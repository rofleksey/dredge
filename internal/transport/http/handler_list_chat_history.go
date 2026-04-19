package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (h *Handler) ListChatHistory(ctx context.Context, params gen.ListChatHistoryParams) (gen.ListChatHistoryRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_chat_history")
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

	list, err := h.twitch.ListChatHistory(ctx, params.Channel, limit)
	if err != nil {
		if errors.Is(err, twitchuc.ErrChannelNotMonitored) {
			return &gen.ErrorMessage{Message: "channel is not monitored"}, nil
		}

		h.obs.LogError(ctx, span, "list chat history failed", err, zap.String("channel", params.Channel))
		return nil, err
	}

	out := make([]gen.ChatHistoryEntry, 0, len(list))

	for _, m := range list {
		out = append(out, chatHistoryEntityToGen(m))
	}

	res := gen.ListChatHistoryOKApplicationJSON(out)

	return &res, nil
}
