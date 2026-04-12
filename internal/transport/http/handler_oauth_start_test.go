package httptransport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"github.com/rofleksey/dredge/internal/service/auth"
	"github.com/rofleksey/dredge/internal/service/settings"
	"github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestHandler_StartTwitchOAuth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	oauth := twitch.NewOAuth(
		"clientid",
		"secret",
		"http://127.0.0.1:8080/oauth/twitch/callback",
		"http://127.0.0.1:5173/#/settings",
		"12345678901234567890123456789012",
	)

	twSvc := twitch.New(repo, noopBroadcaster{}, testTwitchServiceConfig("cid", "sec"), obs)
	setSvc := settings.New(repo, obs)

	h := NewHandler(authSvc, setSvc, twSvc, oauth, obs)

	res, err := h.StartTwitchOAuth(adminCtx(), gen.OptStartTwitchOAuthRequest{})
	require.NoError(t, err)
	assert.Contains(t, res.AuthorizeURL, "https://id.twitch.tv/oauth2/authorize")
}
