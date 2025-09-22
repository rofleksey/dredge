package controller

import (
	"context"
	"dredge/app/api"
	"dredge/pkg/config"

	"github.com/elliotchance/pie/v2"
)

func (s *Server) GetUsers(ctx context.Context, request api.GetUsersRequestObject) (api.GetUsersResponseObject, error) {
	return api.GetUsers200JSONResponse{
		Usernames: pie.Map(s.cfg.Accounts, func(a config.Account) string {
			return a.Username
		}),
	}, nil
}
