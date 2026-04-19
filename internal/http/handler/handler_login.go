package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/http/gen"
	authuc "github.com/rofleksey/dredge/internal/usecase/auth"
)

func (h *Handler) Login(ctx context.Context, req *gen.LoginRequest) (gen.LoginRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.login")
	defer span.End()

	token, err := h.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, authuc.ErrInvalidCredentials) {
			h.obs.Logger.Debug("login failed: invalid credentials")
			return &gen.LoginUnauthorized{}, nil
		}

		h.obs.LogError(ctx, span, "login failed", err)
		return nil, err
	}

	return &gen.LoginResponse{Token: token}, nil
}
