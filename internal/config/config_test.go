package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadErrorOnMissingFile(t *testing.T) {
	_, err := Load("definitely-missing.yaml")
	require.Error(t, err)
}

func TestLoadOK(t *testing.T) {
	path := "test_config.yaml"
	data := []byte(`
server: { address: ":8080", base_url: "http://localhost:5173", metrics_address: ":9090", login_rate_limit_per_minute: 10 }
database: { dsn: "postgres://x" }
jwt: { secret: "1234567890123456", ttl: "1h" }
admin: { email: "owner@example.com", password: "password123" }
twitch: { client_id: "id", client_secret: "secret", oauth_redirect_uri: "http://localhost:8080/oauth/twitch/callback", oauth_return_url: "http://localhost:5173/#/settings" }
observability: { service_name: "dredge-backend", log_level: "info", sentry_dsn: "" }
`)
	err := os.WriteFile(path, data, 0o600)
	require.NoError(t, err)

	defer func() { _ = os.Remove(path) }()

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Server.Address)
}

func TestLoadErrorOnValidation(t *testing.T) {
	path := "test_config_invalid.yaml"
	data := []byte(`
server: { address: ":8080", base_url: "http://localhost:5173" }
database: { dsn: "postgres://x" }
jwt: { secret: "short", ttl: "1h" }
admin: { email: "not-an-email", password: "password123" }
twitch: { client_id: "id", client_secret: "secret", oauth_redirect_uri: "http://localhost:8080/cb", oauth_return_url: "http://localhost:5173/#/settings" }
observability: { service_name: "dredge-backend", log_level: "info", sentry_dsn: "" }
`)
	err := os.WriteFile(path, data, 0o600)
	require.NoError(t, err)

	defer func() { _ = os.Remove(path) }()

	_, err = Load(path)
	require.Error(t, err)
}
