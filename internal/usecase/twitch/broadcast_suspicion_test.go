package twitch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

type captureBroadcaster struct {
	last any
}

func (c *captureBroadcaster) BroadcastJSON(v any) {
	c.last = v
}

func TestPatchTouchesSuspicionFields(t *testing.T) {
	t.Parallel()

	assert.False(t, PatchTouchesSuspicionFields(entity.TwitchUserPatch{Marked: entity.ToPointer(true)}))

	assert.True(t, PatchTouchesSuspicionFields(entity.TwitchUserPatch{IsSus: entity.ToPointer(true)}))
	assert.True(t, PatchTouchesSuspicionFields(entity.TwitchUserPatch{SusType: entity.ToPointer("manual")}))
	assert.True(t, PatchTouchesSuspicionFields(entity.TwitchUserPatch{SusDescription: entity.ToPointer("x")}))
	assert.True(t, PatchTouchesSuspicionFields(entity.TwitchUserPatch{SusAutoSuppressed: entity.ToPointer(true)}))
}

func TestService_BroadcastTwitchUserSuspicion_payload(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	cap := &captureBroadcaster{}

	svc := New(repo, cap, testTwitchCfg("id", "secret"), obs)

	st := "manual"
	sd := "Follows a channel on the blacklist"
	svc.BroadcastTwitchUserSuspicion(entity.TwitchUser{
		ID:             42,
		Username:       "Some_Login",
		IsSus:          true,
		SusType:        &st,
		SusDescription: &sd,
	})

	require.NotNil(t, cap.last)
	m, ok := cap.last.(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "twitch_user_suspicion", m["type"])
	assert.EqualValues(t, 42, m["user_twitch_id"])
	assert.Equal(t, "some_login", m["username"])
	assert.Equal(t, true, m["is_sus"])
	assert.Equal(t, st, m["sus_type"])
	assert.Equal(t, sd, m["sus_description"])
}
