package httptransport

import (
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/service/auth"
	"github.com/rofleksey/dredge/internal/service/settings"
	twitchsvc "github.com/rofleksey/dredge/internal/service/twitch"
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
	twitch      *twitchsvc.Service
	twitchOAuth *twitchsvc.OAuth
	obs         *observability.Stack
}
