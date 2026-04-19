package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/repository/postgres"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	httptransport "github.com/rofleksey/dredge/internal/transport/http"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
	"github.com/rofleksey/dredge/internal/transport/ws"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
	"github.com/rofleksey/dredge/internal/webui"
)

func newPGXPool(cfg config.Config) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}

	if cfg.Database.MaxConns > 0 {
		pcfg.MaxConns = cfg.Database.MaxConns
	}

	if cfg.Database.MinConns > 0 {
		pcfg.MinConns = cfg.Database.MinConns
	}

	return pgxpool.NewWithConfig(context.Background(), pcfg)
}

func fxOptions() fx.Option {
	return fx.Options(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			func() (config.Config, error) { return config.Load("config.yaml") },
			func(cfg config.Config) (config.AllowedWebOrigin, error) {
				return config.ParseAllowedWebOrigin(cfg)
			},
			func(cfg config.Config) (*observability.Stack, error) {
				return observability.Setup(cfg)
			},
			func(obs *observability.Stack) *zap.Logger { return obs.Logger },
			func(obs *observability.Stack) repository.Instrumentation { return obs },
			newPGXPool,
			postgres.New,
			func(r *postgres.Repository) repository.Store { return r },
			func(cfg config.Config, obs *observability.Stack) (*auth.Service, error) {
				return auth.New(cfg, cfg.JWT.Secret, cfg.JWT.TTL, obs)
			},
			settings.New,
			func(origin config.AllowedWebOrigin) (*ws.Hub, error) {
				return ws.NewHub(string(origin)), nil
			},
			func(r repository.Store, hub *ws.Hub, cfg config.Config, obs *observability.Stack) *twitchuc.Service {
				return twitchuc.New(r, hub, cfg, obs)
			},
			func(cfg config.Config) *twitchoauth.OAuth {
				return twitchoauth.NewOAuth(
					cfg.Twitch.ClientID,
					cfg.Twitch.ClientSecret,
					cfg.Twitch.OAuthRedirectURI,
					cfg.Twitch.OAuthReturnURL,
					cfg.JWT.Secret,
				)
			},
			httptransport.NewHandler,
			httptransport.NewSecurity,
			func(cfg config.Config) *httptransport.LoginLimiter {
				return httptransport.NewLoginLimiter(cfg.Server.LoginRateLimitPerMinute)
			},
			func(h *httptransport.Handler, sec *httptransport.Security, limiter *httptransport.LoginLimiter) (*gen.Server, error) {
				return gen.NewServer(h, sec,
					gen.WithMiddleware(httptransport.LoginRateLimitMiddleware(limiter)),
					gen.WithMiddleware(httptransport.RequireAdminMiddleware()),
					gen.WithErrorHandler(httptransport.OgenErrorHandler()),
				)
			},
			func(cfg config.Config, authSvc *auth.Service, srv *gen.Server, hub *ws.Hub, tw *twitchuc.Service, oauth *twitchoauth.OAuth, sett *settings.Service, obs *observability.Stack, origin config.AllowedWebOrigin) (*http.Server, error) {
				mux := http.NewServeMux()
				mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
						return
					}
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("ok"))
				})
				mux.Handle("/ws", httptransport.LiveWebsocketHandler(authSvc, hub, tw, obs.Logger))
				mux.Handle(httptransport.TwitchOAuthCallbackPath, httptransport.NewTwitchOAuthCallback(oauth, sett, obs))
				mux.Handle("/", webui.NewMux(srv))

				return &http.Server{Addr: cfg.Server.Address, Handler: obs.InstrumentHTTP(httptransport.WrapCORS(string(origin), mux))}, nil
			},
		),
		fx.Invoke(registerLifecycle),
	)
}

func New() *fx.App {
	return fx.New(fxOptions())
}
