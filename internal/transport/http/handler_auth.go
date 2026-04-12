package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/service/auth"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) Login(ctx context.Context, req *gen.LoginRequest) (gen.LoginRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.login")
	defer span.End()

	token, err := h.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.obs.Logger.Debug("login failed: invalid credentials")
			return &gen.LoginUnauthorized{}, nil
		}

		h.obs.LogError(ctx, span, "login failed", err)
		return nil, err
	}

	return &gen.LoginResponse{Token: token}, nil
}

func (h *Handler) Me(ctx context.Context) (gen.MeRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.me")
	defer span.End()

	userID, ok := ctx.Value(userIDCtxKey).(int64)
	if !ok {
		return &gen.MeUnauthorized{}, nil
	}

	a, err := h.auth.Me(ctx, userID)
	if err != nil {
		h.obs.LogError(ctx, span, "load me failed", err, zap.Int64("user_id", userID))
		return nil, err
	}

	return &gen.Account{ID: a.ID, Email: a.Email, Role: a.Role}, nil
}
