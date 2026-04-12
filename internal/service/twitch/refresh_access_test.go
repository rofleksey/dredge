package twitch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_refreshAccessToken_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		_ = json.NewEncoder(w).Encode(struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}{AccessToken: "at", RefreshToken: "rt"})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	svc.HTTPClient = srv.Client()
	svc.HTTPClient.Transport = rewriteHostTransport(srv)

	at, rt, err := svc.RefreshAccessToken(context.Background(), "myrefresh")
	require.NoError(t, err)
	assert.Equal(t, "at", at)
	assert.Equal(t, "rt", rt)
}

func TestService_refreshAccessToken_emptyAccessToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}{AccessToken: "", RefreshToken: "r"})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	svc.HTTPClient = srv.Client()
	svc.HTTPClient.Transport = rewriteHostTransport(srv)

	_, _, err := svc.RefreshAccessToken(context.Background(), "rt")
	require.Error(t, err)
}

func TestService_refreshAccessToken_invalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	svc.HTTPClient = srv.Client()
	svc.HTTPClient.Transport = rewriteHostTransport(srv)

	_, _, err := svc.RefreshAccessToken(context.Background(), "rt")
	require.Error(t, err)
}

func TestService_refreshAccessToken_rejected(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	svc.HTTPClient = srv.Client()
	svc.HTTPClient.Transport = rewriteHostTransport(srv)

	_, _, err := svc.RefreshAccessToken(context.Background(), "rt")
	require.Error(t, err)
}

func rewriteHostTransport(srv *httptest.Server) http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		p := strings.TrimPrefix(srv.URL, "http://")
		req.URL.Scheme = "http"
		req.URL.Host = p
		return http.DefaultTransport.RoundTrip(req)
	})
}
