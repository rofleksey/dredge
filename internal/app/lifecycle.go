package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository/postgres"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

type twitchRuntime struct {
	presenceCtx        context.Context
	stopPresence       context.CancelFunc
	streamRecorderCtx  context.Context
	stopStreamRecorder context.CancelFunc
	enrichWorkerCtx    context.Context
	stopEnrichWorker   context.CancelFunc
	persistCtx         context.Context
	stopPersist        context.CancelFunc
	metricsServer      *http.Server
}

func registerLifecycle(
	lc fx.Lifecycle,
	cfg config.Config,
	pool *pgxpool.Pool,
	twitchSvc *twitchuc.Service,
	server *http.Server,
	log *zap.Logger,
	obs *observability.Stack,
) {
	rt := &twitchRuntime{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return onAppStart(ctx, cfg, pool, twitchSvc, server, log, obs, rt)
		},
		OnStop: func(ctx context.Context) error {
			return onAppStop(ctx, pool, twitchSvc, server, log, obs, rt)
		},
	})
}

func onAppStart(
	ctx context.Context,
	cfg config.Config,
	pool *pgxpool.Pool,
	twitchSvc *twitchuc.Service,
	server *http.Server,
	log *zap.Logger,
	obs *observability.Stack,
	rt *twitchRuntime,
) error {
	ctx, startupSpan := obs.Tracer.Start(ctx, "app.startup")
	defer startupSpan.End()

	log.Info("starting dredge backend")

	if err := postgres.RunMigrations(ctx, pool); err != nil {
		startupSpan.RecordError(err)
		sentry.CaptureException(err)
		return err
	}

	rt.presenceCtx, rt.stopPresence = context.WithCancel(context.Background())
	rt.streamRecorderCtx, rt.stopStreamRecorder = context.WithCancel(context.Background())
	rt.enrichWorkerCtx, rt.stopEnrichWorker = context.WithCancel(context.Background())
	rt.persistCtx, rt.stopPersist = context.WithCancel(context.Background())

	twitchSvc.SetPersistContext(rt.persistCtx)

	twitchSvc.StartEnrichmentWorker(rt.enrichWorkerCtx)
	twitchSvc.EnqueueMonitoredAndMarkedUsersForEnrichment(ctx)

	if err := twitchSvc.StartMonitor(ctx); err != nil {
		startupSpan.RecordError(err)
		sentry.CaptureException(err)
		return err
	}

	go twitchSvc.StartPresenceTicker(rt.presenceCtx)
	go twitchSvc.StartStreamSessionRecorder(rt.streamRecorderCtx)

	if addr := cfg.Server.MetricsAddress; addr != "" {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())

		rt.metricsServer = &http.Server{Addr: addr, Handler: metricsMux}

		mln, err := net.Listen("tcp", addr)
		if err != nil {
			startupSpan.RecordError(err)
			sentry.CaptureException(err)
			return err
		}

		go func() {
			if err := rt.metricsServer.Serve(mln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error("metrics server exited", zap.Error(err))
			}
		}()

		log.Info("metrics server listening", zap.String("addr", mln.Addr().String()))
	}

	ln, err := net.Listen("tcp", cfg.Server.Address)
	if err != nil {
		startupSpan.RecordError(err)
		sentry.CaptureException(err)
		return err
	}

	go func() {
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server exited", zap.Error(err))
		}
	}()

	log.Info("http server listening", zap.String("addr", ln.Addr().String()))

	return nil
}

func onAppStop(
	ctx context.Context,
	pool *pgxpool.Pool,
	twitchSvc *twitchuc.Service,
	server *http.Server,
	log *zap.Logger,
	obs *observability.Stack,
	rt *twitchRuntime,
) error {
	log.Info("stopping dredge backend")

	rt.stopPresence()
	rt.stopStreamRecorder()
	rt.stopEnrichWorker()

	twitchSvc.StopMonitor()

	rt.stopPersist()

	if rt.metricsServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = rt.metricsServer.Shutdown(shutdownCtx)
	}

	_ = server.Shutdown(ctx)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		log.Warn("database pool ping at shutdown failed", zap.Error(err))
	}

	pool.Close()
	sentry.Flush(2 * time.Second)

	if obs.LogLoggerProvider != nil {
		_ = obs.LogLoggerProvider.Shutdown(ctx)
	}

	_ = obs.TracerProvider.Shutdown(ctx)
	_ = log.Sync()

	return nil
}
