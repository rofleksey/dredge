package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowedOrigin(t *testing.T) {
	t.Parallel()

	o, err := AllowedOrigin("http://localhost:5173")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:5173", o)

	o, err = AllowedOrigin("https://app.example.com")
	require.NoError(t, err)
	assert.Equal(t, "https://app.example.com", o)
}

func TestParseAllowedWebOrigin(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	cfg.Server.BaseURL = "http://localhost:8080"

	origin, err := ParseAllowedWebOrigin(cfg)
	require.NoError(t, err)
	assert.Equal(t, AllowedWebOrigin("http://localhost:8080"), origin)
}
