package main

import (
	"context"
	"dredge/app/api"
	"dredge/app/controller"
	"dredge/app/service/accounts"
	"dredge/app/service/auth"
	"dredge/app/service/limits"
	"dredge/app/service/messages"
	"dredge/app/service/stalk"
	"dredge/pkg/config"
	"dredge/pkg/database"
	"dredge/pkg/middleware"
	"dredge/pkg/migration"
	"dredge/pkg/routes"
	sentry2 "dredge/pkg/sentry"
	"dredge/pkg/tlog"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do"
)

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	di := do.New()
	do.ProvideValue(di, appCtx)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	do.ProvideValue(di, cfg)

	if err = tlog.Init(cfg); err != nil {
		log.Fatalf("logging init failed: %v", err)
	}

	if err = sentry2.Init(cfg); err != nil {
		slog.Error("Sentry initialization failed", slog.Any("error", err))
	}
	defer sentry.Flush(time.Second)
	defer sentry.RecoverWithContext(appCtx)

	slog.ErrorContext(appCtx, "Service restarted")

	dbConnStr := "postgres://" + cfg.DB.User + ":" + cfg.DB.Pass + "@" + cfg.DB.Host + "/" + cfg.DB.Database + "?sslmode=disable&pool_max_conns=30&pool_min_conns=5&pool_max_conn_lifetime=1h&pool_max_conn_idle_time=30m&pool_health_check_period=1m&connect_timeout=10"

	dbConf, err := pgxpool.ParseConfig(dbConnStr)
	if err != nil {
		log.Fatalf("pgxpool.ParseConfig() failed: %v", err)
	}

	dbConf.ConnConfig.RuntimeParams = map[string]string{
		"statement_timeout":                   "30000",
		"idle_in_transaction_session_timeout": "60000",
	}

	dbConn, err := pgxpool.NewWithConfig(appCtx, dbConf)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	if err = database.InitSchema(appCtx, dbConn); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	do.ProvideValue(di, dbConn)

	queries := database.New(dbConn)
	do.ProvideValue(di, queries)

	if err = migration.Migrate(appCtx, di); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	do.Provide(di, auth.New)
	do.Provide(di, limits.New)
	do.Provide(di, accounts.New)
	do.Provide(di, messages.New)
	do.Provide(di, stalk.New)

	server := controller.NewStrictServer(di)
	handler := api.NewStrictHandler(server, nil)

	app := fiber.New(fiber.Config{
		AppName:          "Dredge API",
		ErrorHandler:     middleware.ErrorHandler,
		ProxyHeader:      "X-Forwarded-For",
		ReadTimeout:      time.Second * 60,
		WriteTimeout:     time.Second * 60,
		DisableKeepalive: false,
	})

	middleware.FiberMiddleware(app, di)
	routes.StaticRoutes(app)

	apiGroup := app.Group("/v1")
	api.RegisterHandlersWithOptions(apiGroup, handler, api.FiberServerOptions{
		BaseURL: "",
		Middlewares: []api.MiddlewareFunc{
			middleware.NewOpenAPIValidator(),
		},
	})

	routes.NotFoundRoute(app)

	go func() {
		defer cancel()

		log.Info("Server started on port 8080")
		if err := app.Listen(":8080"); err != nil {
			log.Warnf("Server stopped! Reason: %v", err)
		}
	}()

	slog.InfoContext(appCtx, "Listening for incoming messages...")
	if err = do.MustInvoke[*stalk.Service](di).Run(); err != nil {
		log.Fatalf("stalk init failed: %v", err)
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Info("Shutting down server...")

		cancel()
	}()

	log.Info("Waiting for services to finish...")
	_ = di.Shutdown()
}
