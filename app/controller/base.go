package controller

import (
	"context"
	"dredge/app/api"
	"dredge/app/service/accounts"
	"dredge/app/service/auth"
	"dredge/app/service/limits"
	"dredge/app/service/messages"
	"dredge/pkg/config"
	"dredge/pkg/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	appCtx          context.Context
	cfg             *config.Config
	dbConn          *pgxpool.Pool
	queries         *database.Queries
	authService     *auth.Service
	accountsService *accounts.Service
	messagesService *messages.Service
	limitsService   *limits.Service
}

func NewStrictServer(di *do.Injector) *Server {
	return &Server{
		appCtx:          do.MustInvoke[context.Context](di),
		cfg:             do.MustInvoke[*config.Config](di),
		dbConn:          do.MustInvoke[*pgxpool.Pool](di),
		queries:         do.MustInvoke[*database.Queries](di),
		authService:     do.MustInvoke[*auth.Service](di),
		accountsService: do.MustInvoke[*accounts.Service](di),
		messagesService: do.MustInvoke[*messages.Service](di),
		limitsService:   do.MustInvoke[*limits.Service](di),
	}
}
