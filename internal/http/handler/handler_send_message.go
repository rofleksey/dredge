package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (h *Handler) SendMessage(ctx context.Context, req *gen.SendMessageRequest) (gen.SendMessageRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.send_message")
	defer span.End()

	err := h.twitch.SendMessage(ctx, req.AccountID, req.Channel, req.Message)
	if err != nil {
		h.obs.LogError(ctx, span, "send message failed", err, zap.Int64("account_id", req.AccountID), zap.String("channel", req.Channel))

		var notice *twitchuc.SendChatNoticeError

		if errors.As(err, &notice) {
			cn := genClientNoticeErr("chat_notice", notice.Message)
			v := gen.SendMessageUnprocessableEntity(cn)

			return &v, nil
		}

		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			cn := genClientNoticeErr("twitch_account_not_found", "Twitch account not found or was removed.")
			v := gen.SendMessageUnprocessableEntity(cn)

			return &v, nil
		}

		if errors.Is(err, twitchuc.ErrSendChatTimeout) {
			cn := genClientNoticeErr("send_chat_timeout", "Timed out sending the message via Twitch.")
			v := gen.SendMessageBadGateway(cn)

			return &v, nil
		}

		cn := genClientNoticeErr("send_failed", err.Error())
		v := gen.SendMessageUnprocessableEntity(cn)

		return &v, nil
	}

	return &gen.SendMessageAccepted{}, nil
}
