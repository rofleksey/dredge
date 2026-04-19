package httpmw

import (
	ogenmw "github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rofleksey/dredge/internal/http/authctx"
	"github.com/rofleksey/dredge/internal/http/gen"
)

// RequireAdminMiddleware enforces role admin for every operation except login and me (Bearer still required for me).
func RequireAdminMiddleware() ogenmw.Middleware {
	return func(req ogenmw.Request, next ogenmw.Next) (ogenmw.Response, error) {
		switch req.OperationName {
		case gen.LoginOperation, gen.MeOperation:
			return next(req)
		default:
			if role, ok := authctx.Role(req.Context); !ok || role != "admin" {
				return ogenmw.Response{}, ogenerrors.ErrSecurityRequirementIsNotSatisfied
			}

			return next(req)
		}
	}
}
