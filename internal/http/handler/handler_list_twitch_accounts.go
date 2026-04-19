package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) ListTwitchAccounts(ctx context.Context) ([]gen.TwitchAccount, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_accounts")
	defer span.End()

	list, err := h.sett.ListTwitchAccounts(ctx)
	if err != nil {
		h.obs.LogError(ctx, span, "list twitch accounts failed", err)
		return nil, err
	}

	out := make([]gen.TwitchAccount, 0, len(list))

	for _, a := range list {
		out = append(out, twitchAccountToAPI(a))
	}

	return out, nil
}
