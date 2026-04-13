package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Address string `yaml:"address" validate:"required"`
		// BaseURL is the public origin of the web UI (scheme + host + port, no path), used for CORS and WebSocket Origin checks.
		BaseURL string `yaml:"base_url" validate:"required,url"`
		// MetricsAddress, if non-empty, is the listen address for Prometheus metrics only (e.g. ":9090"). The main Server.Address does not expose /metrics.
		MetricsAddress string `yaml:"metrics_address" validate:"omitempty"`
		// LoginRateLimitPerMinute caps POST /auth/login per client IP per rolling minute; 0 disables limiting.
		LoginRateLimitPerMinute int `yaml:"login_rate_limit_per_minute" validate:"min=0"`
	} `yaml:"server" validate:"required"`
	Database struct {
		DSN string `yaml:"dsn" validate:"required"`
		// MaxConns and MinConns are passed to pgxpool when > 0 (otherwise pool defaults apply).
		MaxConns int32 `yaml:"max_conns" validate:"omitempty,min=1"`
		MinConns int32 `yaml:"min_conns" validate:"omitempty,min=0"`
	} `yaml:"database" validate:"required"`
	JWT struct {
		Secret string        `yaml:"secret" validate:"required,min=16"`
		TTL    time.Duration `yaml:"ttl" validate:"required"`
	} `yaml:"jwt" validate:"required"`
	Admin struct {
		Email    string `yaml:"email" validate:"required,email"`
		Password string `yaml:"password" validate:"required,min=8"`
	} `yaml:"admin" validate:"required"`
	Twitch struct {
		ClientID     string `yaml:"client_id" validate:"required"`
		ClientSecret string `yaml:"client_secret" validate:"required"`
		// OAuthRedirectURI is the exact URL registered in the Twitch developer console (e.g. https://api.example.com/oauth/twitch/callback).
		OAuthRedirectURI string `yaml:"oauth_redirect_uri" validate:"required,url"`
		// OAuthReturnURL is where the user is sent after linking (SPA settings route with hash routing, e.g. http://localhost:5173/#/settings).
		OAuthReturnURL string `yaml:"oauth_return_url" validate:"required,url"`
		// EnrichmentCron is a robfig/cron spec (UTC) for daily Helix account/follow enrichment. Default 0 3 * * *.
		EnrichmentCron string `yaml:"enrichment_cron"`
		// ViewerPollInterval is how often the watch UI should refresh Helix stream metadata (viewer count, live state). Default 10s.
		ViewerPollInterval time.Duration `yaml:"viewer_poll_interval"`
		// ChannelChattersSyncInterval is how often IRC NAMES lists are merged into channel_chatters. Default 10s.
		ChannelChattersSyncInterval time.Duration `yaml:"channel_chatters_sync_interval"`
		// StreamSessionPollInterval is how often Helix is polled to open/close persisted stream sessions for monitored channels. Default: same as viewer_poll_interval.
		StreamSessionPollInterval time.Duration `yaml:"stream_session_poll_interval"`
		// UserOAuthTokenCacheTTL is how long a linked-account OAuth access token is reused before refresh (Helix + IRC). Default 30m.
		UserOAuthTokenCacheTTL time.Duration `yaml:"user_oauth_token_cache_ttl"`
	} `yaml:"twitch" validate:"required"`
	Observability struct {
		ServiceName   string `yaml:"service_name" validate:"required"`
		LogLevel      string `yaml:"log_level" validate:"omitempty,oneof=debug info warn error"`
		SentryDSN     string `yaml:"sentry_dsn"`
		TraceExporter string `yaml:"trace_exporter" validate:"omitempty,oneof=stdout none"`
		// LogExporter: none — console only (human-readable). otlp — also export logs via OTLP (HTTP); use OTEL_EXPORTER_OTLP_* env vars.
		LogExporter string `yaml:"log_exporter" validate:"omitempty,oneof=none otlp"`
	} `yaml:"observability" validate:"required"`
}

func Load(path string) (Config, error) {
	if _, err := os.Stat(path); err != nil {
		return Config{}, fmt.Errorf("config must exist at %q: %w", path, err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	if cfg.Twitch.ViewerPollInterval <= 0 {
		cfg.Twitch.ViewerPollInterval = 10 * time.Second
	}

	if cfg.Twitch.ChannelChattersSyncInterval <= 0 {
		cfg.Twitch.ChannelChattersSyncInterval = 10 * time.Second
	}

	if cfg.Twitch.StreamSessionPollInterval <= 0 {
		cfg.Twitch.StreamSessionPollInterval = cfg.Twitch.ViewerPollInterval
	}

	if cfg.Twitch.UserOAuthTokenCacheTTL <= 0 {
		cfg.Twitch.UserOAuthTokenCacheTTL = 30 * time.Minute
	}

	return cfg, nil
}
