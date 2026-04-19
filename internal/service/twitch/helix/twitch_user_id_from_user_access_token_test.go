package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"

	"go.uber.org/mock/gomock"
)

func TestTwitchUserIDFromUserAccessToken_ok(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/helix/users", r.URL.Path)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{{"id": "141981764"}},
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	c := NewClient(repo, obs, "cid", "csec")
	c.HTTPClient = srv.Client()
	c.HTTPClient.Transport = roundTripRewriteHost(srv)

	id, err := c.TwitchUserIDFromUserAccessToken(context.Background(), "user-token")
	require.NoError(t, err)
	require.Equal(t, int64(141981764), id)
}
