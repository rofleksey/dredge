package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/rofleksey/dredge/internal/config"
)

func TestNew_returnsFxApp(t *testing.T) {
	t.Parallel()

	a := New()
	require.NotNil(t, a)
}

func TestFx_options_validateGraph(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join("..", "..", "config.example.yaml")
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err, "load %s (run tests from module root)", cfgPath)

	require.NoError(t, fx.ValidateApp(fx.Options(
		fxOptions(),
		fx.Replace(cfg),
	)))
}
