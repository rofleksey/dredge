package httptransport

import (
	"github.com/rofleksey/dredge/internal/observability"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

type contextKey string

const (
	userIDCtxKey contextKey = "user_id"
	roleCtxKey   contextKey = "role"
)

type Security struct {
	auth *auth.Service
	obs  *observability.Stack
}

type Handler struct {
	auth        *auth.Service
	sett        *settings.Service
	twitch      *twitchuc.Service
	twitchOAuth *twitchoauth.OAuth
	obs         *observability.Stack
}
