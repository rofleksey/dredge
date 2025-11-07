package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/samber/oops"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Service name for telemetry and logs
	ServiceName string    `yaml:"service_name" env:"SERVICE_NAME" example:"dredge" validate:"required"`
	Sentry      Sentry    `yaml:"sentry" envPrefix:"SENTRY_"`
	Log         Log       `yaml:"log" envPrefix:"LOG_"`
	Telemetry   Telemetry `yaml:"telemetry" envPrefix:"TELEMETRY_"`
	Twitch      Twitch    `yaml:"twitch" envPrefix:"TWITCH_"`
	Alert       Alert     `yaml:"alert" envPrefix:"ALERT_"`
}

type Sentry struct {
	DSN string `yaml:"dsn" env:"DSN" example:"https://a1b2c3d4e5f6g7h8a1b2c3d4e5f6g7h8@o123456.ingest.sentry.io/1234567"`
}

type Log struct {
	// Telegram logging config
	Telegram TelegramLog `yaml:"telegram" envPrefix:"TELEGRAM_"`
}

type TelegramLog struct {
	// Chat bot token, obtain it via BotFather
	Token string `yaml:"token" env:"TOKEN" example:"1234567890:ABCdefGHIjklMNopQRstUVwxyZ-123456789"`
	// Chat ID to send messages to
	ChatID string `yaml:"chat_id" env:"CHAT_ID" example:"1001234567890"`
}

type Telemetry struct {
	// Whether to enable opentelemetry logs/metrics/traces export
	Enabled bool `yaml:"enabled" env:"ENABLED" example:"false"`
}

type Twitch struct {
	// ClientID of the twitch_api application
	ClientID string `yaml:"client_id" example:"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p" validate:"required"`
	// Client secret of the twitch_api application
	ClientSecret string `yaml:"client_secret" example:"abc123def456ghi789jkl012mno345pqr678stu901" validate:"required"`
	// Username of the bot account
	Username string `yaml:"username" example:"PogChamp123" validate:"required"`
	// User refresh token of the bot account
	RefreshToken string `yaml:"refresh_token" example:"v1.abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567" validate:"required"`
}

type AlertEntry struct {
	ExcludeChannels  []string `json:"exclude_channels"`
	ExcludeUsernames []string `json:"exclude_usernames"`
	Message          string   `json:"message"`
}

type Alert struct {
	// List of alert entries
	List []AlertEntry `yaml:"list" validate:"required"`
	// List of usernames to exclude
	ExcludeUsernames []string `yaml:"exclude_usernames"`
	// List of channels to monitor even if they are offline
	PermanentChannels []string `yaml:"permanent_channels"`
}

func Load(configPath string) (*Config, error) {
	var result Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, oops.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, oops.Errorf("failed to parse YAML config: %w", err)
	}

	if err := env.ParseWithOptions(&result, env.Options{ //nolint:exhaustruct
		Prefix: "DREDGE_",
	}); err != nil {
		return nil, oops.Errorf("failed to parse environment variables: %w", err)
	}

	if result.ServiceName == "" {
		result.ServiceName = "dredge"
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(result); err != nil {
		return nil, oops.Errorf("failed to validate config: %w", err)
	}

	return &result, nil
}
