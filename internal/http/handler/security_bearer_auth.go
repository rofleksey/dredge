package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/authctx"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (s *Security) HandleBearerAuth(ctx context.Context, _ gen.OperationName, t gen.BearerAuth) (context.Context, error) {
	ctx, span := s.obs.StartSpan(ctx, "handler.security.bearer")
	defer span.End()

	userID, role, err := s.auth.ParseToken(ctx, t.Token)
	if err != nil {
		s.obs.LogError(ctx, span, "bearer auth failed", err)
		return nil, err
	}

	ctx = authctx.WithUserID(ctx, userID)
	ctx = authctx.WithRole(ctx, role)

	return ctx, nil
}
