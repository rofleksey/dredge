package handler

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/http/authctx"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) Me(ctx context.Context) (gen.MeRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.me")
	defer span.End()

	userID, ok := authctx.UserID(ctx)
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
