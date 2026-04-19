package httptransport

import (
	"errors"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rofleksey/dredge/internal/observability"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func NewSecurity(a *auth.Service, obs *observability.Stack) *Security {
	return &Security{auth: a, obs: obs}
}

func NewHandler(a *auth.Service, sett *settings.Service, t *twitchuc.Service, oauth *twitchoauth.OAuth, obs *observability.Stack) *Handler {
	return &Handler{auth: a, sett: sett, twitch: t, twitchOAuth: oauth, obs: obs}
}

var _ gen.Handler = (*Handler)(nil)
var _ gen.SecurityHandler = (*Security)(nil)

func IsUnauthorized(err error) bool {
	return errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied)
}
