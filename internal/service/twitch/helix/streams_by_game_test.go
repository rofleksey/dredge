package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestHelixStreamsByGameIDPage_ok(t *testing.T) {
	t.Parallel()

	var calls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		switch {
		case strings.Contains(r.URL.Path, "oauth2/token"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at", "expires_in": 3600})

		case strings.HasPrefix(r.URL.Path, "/helix/streams"):
			require.Equal(t, "42", r.URL.Query().Get("game_id"))
			require.Equal(t, "100", r.URL.Query().Get("first"))
			require.Equal(t, "abc", r.URL.Query().Get("after"))

			_, _ = w.Write([]byte(`{
  "data": [
    {
      "user_id": "9",
      "user_login": "SomeOne",
      "title": "hi",
      "game_name": "G",
      "viewer_count": 12,
      "tags": ["English", "FPS"]
    }
  ],
  "pagination": { "cursor": "nextcur" }
}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	c := NewClient(repo, obs, "cid", "csec")
	c.HTTPClient = srv.Client()
	c.HTTPClient.Transport = roundTripRewriteHost(srv)

	rows, next, err := c.HelixStreamsByGameIDPage(context.Background(), "42", 100, "abc")
	require.NoError(t, err)
	require.Equal(t, "nextcur", next)
	require.Len(t, rows, 1)
	require.Equal(t, int64(9), rows[0].UserID)
	require.Equal(t, "someone", rows[0].UserLogin)
	require.Equal(t, int64(12), rows[0].ViewerCount)
	require.Equal(t, "hi", rows[0].Title)
	require.Equal(t, "G", rows[0].GameName)
	require.Equal(t, []string{"English", "FPS"}, rows[0].Tags)
	require.Equal(t, 2, calls)
}

func TestHelixStreamsByGameIDPage_tagsOmitted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "oauth2/token"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at", "expires_in": 3600})

		case strings.HasPrefix(r.URL.Path, "/helix/streams"):
			_, _ = w.Write([]byte(`{
  "data": [
    {
      "user_id": "1",
      "user_login": "a",
      "title": "",
      "game_name": "",
      "viewer_count": 0
    }
  ],
  "pagination": {}
}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	c := NewClient(repo, obs, "cid", "csec")
	c.HTTPClient = srv.Client()
	c.HTTPClient.Transport = roundTripRewriteHost(srv)

	rows, next, err := c.HelixStreamsByGameIDPage(context.Background(), "1", 20, "")
	require.NoError(t, err)
	require.Empty(t, next)
	require.Len(t, rows, 1)
	require.Empty(t, rows[0].Tags)
}

func TestHelixStreamsByGameIDPage_emptyGameID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	c := NewClient(repo, obs, "cid", "csec")

	_, _, err := c.HelixStreamsByGameIDPage(context.Background(), "  ", 10, "")
	require.ErrorIs(t, err, ErrInvalidGameID)
}
