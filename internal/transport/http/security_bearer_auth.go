package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (s *Security) HandleBearerAuth(ctx context.Context, _ gen.OperationName, t gen.BearerAuth) (context.Context, error) {
	ctx, span := s.obs.StartSpan(ctx, "handler.security.bearer")
	defer span.End()

	userID, role, err := s.auth.ParseToken(ctx, t.Token)
	if err != nil {
		s.obs.LogError(ctx, span, "bearer auth failed", err)
		return nil, err
	}

	ctx = context.WithValue(ctx, userIDCtxKey, userID)
	ctx = context.WithValue(ctx, roleCtxKey, role)

	return ctx, nil
}
