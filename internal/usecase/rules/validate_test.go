package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestValidateRule_interval_ok(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		Name:           "interval",
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
		Name:           "x",
		EventType:      EventInterval,
		EventSettings:  map[string]any{"interval_seconds": 30.0},
		ActionType:     ActionNotify,
		ActionSettings: map[string]any{},
	}

	err := ValidateRule(r)
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}

func TestValidateRule_send_chat_ok(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		Name:           "sc",
		EventType:      EventChatMessage,
		EventSettings:  map[string]any{},
		Middlewares:    nil,
		ActionType:     ActionSendChat,
		ActionSettings: map[string]any{"message": "hello $CHANNEL"},
	}

	require.NoError(t, ValidateRule(r))
}

func TestValidateRule_send_chat_missing_message(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		Name:           "sc",
		EventType:      EventChatMessage,
		EventSettings:  map[string]any{},
		ActionType:     ActionSendChat,
		ActionSettings: map[string]any{},
	}

	err := ValidateRule(r)
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}

func TestValidateRule_name_empty(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		EventType:      EventChatMessage,
		EventSettings:  map[string]any{},
		ActionType:     ActionNotify,
		ActionSettings: map[string]any{},
	}

	err := ValidateRule(r)
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}
