package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListTwitchAccounts(ctx context.Context) ([]gen.TwitchAccount, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_accounts")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

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

func (h *Handler) CreateTwitchAccount(ctx context.Context, req *gen.CreateTwitchAccountRequest) (*gen.TwitchAccount, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.create_twitch_account")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

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

func (h *Handler) UpdateTwitchAccount(ctx context.Context, req *gen.UpdateTwitchAccountPostRequest) (gen.UpdateTwitchAccountRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_twitch_account")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	at := string(req.GetAccountType())

	a, err := h.sett.PatchTwitchAccount(ctx, req.GetID(), &at)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return &gen.ErrorMessage{Message: "twitch account not found"}, nil
		}

		h.obs.LogError(ctx, span, "update twitch account failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	out := twitchAccountToAPI(a)

	return &out, nil
}

func (h *Handler) DeleteTwitchAccount(ctx context.Context, req *gen.DeleteByIDRequest) (gen.DeleteTwitchAccountRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.delete_twitch_account")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	if err := h.sett.DeleteTwitchAccount(ctx, req.ID); err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return &gen.ErrorMessage{Message: "twitch account not found"}, nil
		}

		h.obs.LogError(ctx, span, "delete twitch account failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	return &gen.DeleteTwitchAccountNoContent{}, nil
}

func (h *Handler) StartTwitchOAuth(ctx context.Context, req gen.OptStartTwitchOAuthRequest) (*gen.StartTwitchOAuthResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.start_twitch_oauth")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	spaReturn := ""

	if req.IsSet() {
		body, _ := req.Get()
		if v, ok := body.GetReturnURL().Get(); ok {
			spaReturn = v
		}
	}

	state, err := h.twitchOAuth.NewState(spaReturn)
	if err != nil {
		h.obs.LogError(ctx, span, "twitch oauth state failed", err)
		return nil, err
	}

	return &gen.StartTwitchOAuthResponse{AuthorizeURL: h.twitchOAuth.AuthorizeURL(state)}, nil
}
