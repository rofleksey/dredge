package httptransport

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) CreateTwitchAccount(ctx context.Context, req *gen.CreateTwitchAccountRequest) (*gen.TwitchAccount, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.create_twitch_account")
	defer span.End()

	accountType := "main"
	if req.AccountType.IsSet() {
		accountType = string(req.AccountType.Value)
	}

	a, err := h.sett.CreateTwitchAccount(ctx, req.GetID(), req.Username, req.RefreshToken, accountType)
	if err != nil {
		h.obs.LogError(ctx, span, "create twitch account failed", err, zap.String("username", req.Username))
		return nil, err
	}

	out := twitchAccountToAPI(a)

	return &out, nil
}
