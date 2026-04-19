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

func TestSendChatMessage_ok(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/helix/chat/messages", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"is_sent": true}},
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

	err := c.SendChatMessage(context.Background(), "tok", 12826, 141981764, "hi")
	require.NoError(t, err)
}

func TestSendChatMessage_helixMessageError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"missing scope"}`))
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	c := NewClient(repo, obs, "cid", "csec")
	c.HTTPClient = srv.Client()
	c.HTTPClient.Transport = roundTripRewriteHost(srv)

	err := c.SendChatMessage(context.Background(), "tok", 1, 2, "hi")
	require.Error(t, err)

	var ce *ChatSendError
	require.ErrorAs(t, err, &ce)
	require.Equal(t, "missing scope", ce.Message)
}
