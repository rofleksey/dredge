package handler

import (
	"github.com/rofleksey/dredge/internal/observability"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/usecase/ai"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/rules"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	"github.com/rofleksey/dredge/internal/usecase/stats"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

type Security struct {
	auth *auth.Usecase
	obs  *observability.Stack
}

type Handler struct {
	auth        *auth.Usecase
	sett        *settings.Usecase
	rules       *rules.Usecase
	twitch      *twitchuc.Usecase
	twitchOAuth *twitchoauth.OAuth
	ai          *ai.Usecase
	stats       *stats.Collector
	obs         *observability.Stack
}
