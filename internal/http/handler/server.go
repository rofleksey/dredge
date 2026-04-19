package handler

import (
	"errors"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/observability"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/usecase/ai"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/rules"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func NewSecurity(a *auth.Usecase, obs *observability.Stack) *Security {
	return &Security{auth: a, obs: obs}
}

func NewHandler(a *auth.Usecase, sett *settings.Usecase, rulesSvc *rules.Usecase, t *twitchuc.Usecase, oauth *twitchoauth.OAuth, aiSvc *ai.Usecase, obs *observability.Stack) *Handler {
	return &Handler{auth: a, sett: sett, rules: rulesSvc, twitch: t, twitchOAuth: oauth, ai: aiSvc, obs: obs}
}

var _ gen.Handler = (*Handler)(nil)
var _ gen.SecurityHandler = (*Security)(nil)

func IsUnauthorized(err error) bool {
	return errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied)
}
