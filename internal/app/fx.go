package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/http/handler"
	httpmw "github.com/rofleksey/dredge/internal/http/middleware"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	"github.com/rofleksey/dredge/internal/repository/postgres"
	twitchoauth "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/usecase/ai"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/rules"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	"github.com/rofleksey/dredge/internal/usecase/stats"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
	"github.com/rofleksey/dredge/internal/webui"
	"github.com/rofleksey/dredge/internal/ws"
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
			func(cfg config.Config, obs *observability.Stack) (*auth.Usecase, error) {
				return auth.New(cfg, cfg.JWT.Secret, cfg.JWT.TTL, obs)
			},
			settings.New,
			func(origin config.AllowedWebOrigin) (*ws.Hub, error) {
				return ws.NewHub(string(origin)), nil
			},
			func(r repository.Store, hub *ws.Hub, cfg config.Config, obs *observability.Stack) *twitchuc.Usecase {
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
			newRulesServices,
			func(r repository.Store, tw *twitchuc.Usecase, rulesSvc *rules.Usecase, sett *settings.Usecase, hub *ws.Hub, obs *observability.Stack) *ai.Usecase {
				return ai.New(r, tw, rulesSvc, sett, hub, obs)
			},
			func(r repository.Store, pool *pgxpool.Pool, tw *twitchuc.Usecase, lim *httpmw.LoginLimiter) *stats.Collector {
				return stats.NewCollector(r, tw, lim, pool)
			},
			handler.NewHandler,
			handler.NewSecurity,
			func(cfg config.Config) *httpmw.LoginLimiter {
				return httpmw.NewLoginLimiter(cfg.Server.LoginRateLimitPerMinute)
			},
			func(h *handler.Handler, sec *handler.Security, limiter *httpmw.LoginLimiter) (*gen.Server, error) {
				return gen.NewServer(h, sec,
					gen.WithMiddleware(httpmw.LoginRateLimitMiddleware(limiter)),
					gen.WithMiddleware(httpmw.RequireAdminMiddleware()),
					gen.WithErrorHandler(httpmw.OgenErrorHandler()),
				)
			},
			func(cfg config.Config, authSvc *auth.Usecase, srv *gen.Server, hub *ws.Hub, tw *twitchuc.Usecase, oauth *twitchoauth.OAuth, sett *settings.Usecase, obs *observability.Stack, origin config.AllowedWebOrigin) (*http.Server, error) {
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
				mux.Handle("/ws", handler.LiveWebsocketHandler(authSvc, hub, tw, obs.Logger))
				mux.Handle(handler.TwitchOAuthCallbackPath, handler.NewTwitchOAuthCallback(oauth, sett, obs))
				mux.Handle("/api/v1/", srv)
				mux.Handle("/", webui.NewMux())

				chain := httpmw.WrapSecurityHeaders(httpmw.WrapCORS(string(origin), mux))

				return &http.Server{Addr: cfg.Server.Address, Handler: obs.InstrumentHTTP(chain)}, nil
			},
		),
		// registerLifecycle must run first: RunMigrations runs in its OnStart before any
		// code (e.g. rules Bootstrap listing rules) that depends on the current schema.
		fx.Invoke(registerLifecycle),
		fx.Invoke(registerRulesLifecycle),
	)
}

func New() *fx.App {
	return fx.New(
		fxOptions(),
		fx.StopTimeout(45*time.Second),
	)
}
