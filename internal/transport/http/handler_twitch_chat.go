package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	twitchsvc "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
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
		if errors.Is(err, twitchsvc.ErrChannelNotMonitored) {
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

func (h *Handler) ListTwitchMessages(ctx context.Context, params gen.ListTwitchMessagesParams) ([]gen.ChatHistoryEntry, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_messages")
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

	if params.Channel.IsSet() {
		f.Channel = params.Channel.Value
	}

	if t, ok := params.CreatedFrom.Get(); ok {
		f.CreatedFrom = &t
	}

	if t, ok := params.CreatedTo.Get(); ok {
		f.CreatedTo = &t
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

	list, err := h.twitch.ListChatMessages(ctx, f)
	if err != nil {
		h.obs.LogError(ctx, span, "list twitch messages failed", err)
		return nil, err
	}

	out := make([]gen.ChatHistoryEntry, 0, len(list))
	for _, m := range list {
		out = append(out, chatHistoryEntityToGen(m))
	}

	return out, nil
}
