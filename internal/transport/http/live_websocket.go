package httptransport

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/transport/ws"
	"github.com/rofleksey/dredge/internal/usecase/auth"
)

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

		if role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
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
