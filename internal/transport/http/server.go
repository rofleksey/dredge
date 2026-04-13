package httptransport

import (
	"context"
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/ogenerrors"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/service/auth"
	"github.com/rofleksey/dredge/internal/service/settings"
	twitchsvc "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
	"github.com/rofleksey/dredge/internal/transport/ws"
)

func NewSecurity(a *auth.Service, obs *observability.Stack) *Security {
	return &Security{auth: a, obs: obs}
}

func (s *Security) HandleBearerAuth(ctx context.Context, _ gen.OperationName, t gen.BearerAuth) (context.Context, error) {
	ctx, span := s.obs.StartSpan(ctx, "handler.security.bearer")
	defer span.End()

	userID, role, err := s.auth.ParseToken(ctx, t.Token)
	if err != nil {
		s.obs.LogError(ctx, span, "bearer auth failed", err)
		return nil, err
	}

	ctx = context.WithValue(ctx, userIDCtxKey, userID)
	ctx = context.WithValue(ctx, roleCtxKey, role)

	return ctx, nil
}

func NewHandler(a *auth.Service, sett *settings.Service, t *twitchsvc.Service, oauth *twitchsvc.OAuth, obs *observability.Stack) *Handler {
	return &Handler{auth: a, sett: sett, twitch: t, twitchOAuth: oauth, obs: obs}
}

func requireAdmin(ctx context.Context) error {
	if role, _ := ctx.Value(roleCtxKey).(string); role != "admin" {
		return ogenerrors.ErrSecurityRequirementIsNotSatisfied
	}
	return nil
}

// LiveWebsocketWelcomer supplies optional JSON payloads queued to the client right after a successful WS upgrade.
type LiveWebsocketWelcomer interface {
	LiveWebSocketWelcomePayloads(ctx context.Context) ([]any, error)
}

func LiveWebsocketHandler(authSvc *auth.Service, hub *ws.Hub, welcomer LiveWebsocketWelcomer, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string

		const prefix = "Bearer "

		raw := r.Header.Get("Authorization")
		if len(raw) > len(prefix) && raw[:len(prefix)] == prefix {
			token = raw[len(prefix):]
		}

		if token == "" {
			// Intentional: query token supports WebSocket clients that cannot set custom headers
			// (e.g. browser WebSocket API). Callers should treat URLs with ?token= as sensitive.
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		userID, role, err := authSvc.ParseToken(r.Context(), token)
		if err != nil || userID <= 0 || role == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		var initial []any

		if welcomer != nil {
			var err error

			initial, err = welcomer.LiveWebSocketWelcomePayloads(r.Context())
			if err != nil {
				if log != nil {
					log.Debug("websocket welcome payloads failed", zap.Error(err))
				}

				initial = nil
			}
		}

		if err := hub.Upgrade(w, r, userID, initial...); err != nil {
			if log != nil {
				log.Debug("websocket upgrade failed", zap.Error(err), zap.String("origin", r.Header.Get("Origin")))
			}
			return
		}
	}
}

var _ gen.Handler = (*Handler)(nil)
var _ gen.SecurityHandler = (*Security)(nil)

func IsUnauthorized(err error) bool {
	return errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied)
}
