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

func TestValidateRule_send_chat_ok_with_account_id(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		Name:           "sc",
		EventType:      EventChatMessage,
		EventSettings:  map[string]any{},
		Middlewares:    nil,
		ActionType:     ActionSendChat,
		ActionSettings: map[string]any{"message": "hello $CHANNEL", "account_id": float64(42)},
	}

	require.NoError(t, ValidateRule(r))
}

func TestValidateRule_send_chat_invalid_account_id(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		Name:           "sc",
		EventType:      EventChatMessage,
		EventSettings:  map[string]any{},
		Middlewares:    nil,
		ActionType:     ActionSendChat,
		ActionSettings: map[string]any{"message": "hello", "account_id": float64(1.5)},
	}

	err := ValidateRule(r)
	require.Error(t, err)
	require.ErrorIs(t, err, entity.ErrInvalidRule)
}

func TestParseSendChatAccountID(t *testing.T) {
	t.Parallel()

	id, err := ParseSendChatAccountID(nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), id)

	id, err = ParseSendChatAccountID(map[string]any{})
	require.NoError(t, err)
	require.Equal(t, int64(0), id)

	id, err = ParseSendChatAccountID(map[string]any{"account_id": float64(7)})
	require.NoError(t, err)
	require.Equal(t, int64(7), id)

	id, err = ParseSendChatAccountID(map[string]any{"account_id": "99"})
	require.NoError(t, err)
	require.Equal(t, int64(99), id)

	_, err = ParseSendChatAccountID(map[string]any{"account_id": float64(-1)})
	require.Error(t, err)
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
