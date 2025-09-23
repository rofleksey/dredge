package controller

import (
	"context"
	"dredge/app/api"
	"dredge/pkg/config"
	"net/http"

	"github.com/elliotchance/pie/v2"
	"github.com/samber/oops"
)

func (s *Server) GetUsers(ctx context.Context, request api.GetUsersRequestObject) (api.GetUsersResponseObject, error) {
	if !s.authService.IsAdmin(ctx) {
		return nil, oops.With("statusCode", http.StatusUnauthorized).New("Unauthorized")
	}

	return api.GetUsers200JSONResponse{
		Usernames: pie.Map(s.cfg.Accounts, func(a config.Account) string {
			return a.Username
		}),
	}, nil
}
