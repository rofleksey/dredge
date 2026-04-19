package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestValidateRule_interval_ok(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		Enabled:        true,
		EventType:      EventInterval,
		EventSettings:  map[string]any{"interval_seconds": 30.0, "channel": "foo"},
		Middlewares:    nil,
		ActionType:     ActionNotify,
		ActionSettings: map[string]any{},
		UseSharedPool:  true,
	}

	require.NoError(t, ValidateRule(r))
}

func TestValidateRule_interval_missing_channel(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		EventType:      EventInterval,
		EventSettings:  map[string]any{"interval_seconds": 30.0},
		ActionType:     ActionNotify,
		ActionSettings: map[string]any{},
	}

	err := ValidateRule(r)
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}
