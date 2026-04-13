package httptransport

import (
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

// RequireAdminMiddleware enforces role admin for every operation except login and me (Bearer still required for me).
func RequireAdminMiddleware() middleware.Middleware {
	return func(req middleware.Request, next middleware.Next) (middleware.Response, error) {
		switch req.OperationName {
		case gen.LoginOperation, gen.MeOperation:
			return next(req)
		default:
			if role, _ := req.Context.Value(roleCtxKey).(string); role != "admin" {
				return middleware.Response{}, ogenerrors.ErrSecurityRequirementIsNotSatisfied
			}
			return next(req)
		}
	}
}
