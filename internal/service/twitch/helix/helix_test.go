package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestResolveChannel_invalidName(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	c := NewClient(repo, obs, "c", "s")

	_, err := c.ResolveChannel(context.Background(), "bad")
	require.ErrorIs(t, err, ErrInvalidChannelName)

	_, err = c.ResolveChannel(context.Background(), "")
	require.ErrorIs(t, err, ErrInvalidChannelName)
}

func TestCachedUserAccessTokenForAccount_reusesWithinTTL(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)

		_ = json.NewEncoder(w).Encode(refreshResp{
			AccessToken: "at1",
			ExpiresIn:   3600,
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

	ctx := context.Background()

	at1, newRT1, err := c.CachedUserAccessTokenForAccount(ctx, 42, "refresh-1")
	require.NoError(t, err)
	require.Equal(t, "at1", at1)
	require.Equal(t, "", newRT1)

	at2, newRT2, err := c.CachedUserAccessTokenForAccount(ctx, 42, "refresh-1")
	require.NoError(t, err)
	require.Equal(t, "at1", at2)
	require.Equal(t, "", newRT2)

	require.EqualValues(t, 1, calls.Load(), "oauth refresh should run once when cache is warm")
}

func TestCachedUserAccessTokenForAccount_refreshTokenChangeBustsCache(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)

		_ = json.NewEncoder(w).Encode(refreshResp{
			AccessToken: "at",
			ExpiresIn:   3600,
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

	ctx := context.Background()

	_, _, err := c.CachedUserAccessTokenForAccount(ctx, 7, "rt-a")
	require.NoError(t, err)
	_, _, err = c.CachedUserAccessTokenForAccount(ctx, 7, "rt-b")
	require.NoError(t, err)

	require.EqualValues(t, 2, calls.Load())
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func roundTripRewriteHost(srv *httptest.Server) http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		p := strings.TrimPrefix(srv.URL, "http://")
		req.URL.Scheme = "http"
		req.URL.Host = p
		return http.DefaultTransport.RoundTrip(req)
	})
}
