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
)

func TestOAuth_ExchangeCode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at",
			"refresh_token": "rt",
			"expires_in":    3600,
		})
	}))
	defer srv.Close()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	o.httpClient = srv.Client()
	o.redirectURI = "http://localhost/cb"

	orig := o.httpClient.Transport

	t.Cleanup(func() { o.httpClient.Transport = orig })

	o.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	})

	out, err := o.ExchangeCode(context.Background(), "code123")
	require.NoError(t, err)
	assert.Equal(t, "at", out.AccessToken)
	assert.Equal(t, "rt", out.RefreshToken)
}

func TestOAuth_ExchangeCode_badStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("no"))
	}))
	defer srv.Close()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	o.httpClient = srv.Client()
	o.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		p := strings.TrimPrefix(srv.URL, "http://")
		req.URL.Scheme = "http"
		req.URL.Host = p
		return http.DefaultTransport.RoundTrip(req)
	})

	_, err := o.ExchangeCode(context.Background(), "code")
	require.Error(t, err)
}

func TestOAuth_ExchangeCode_emptyRefreshToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at",
			"refresh_token": "",
			"expires_in":    3600,
		})
	}))
	defer srv.Close()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	o.httpClient = srv.Client()
	o.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		p := strings.TrimPrefix(srv.URL, "http://")
		req.URL.Scheme = "http"
		req.URL.Host = p
		return http.DefaultTransport.RoundTrip(req)
	})

	_, err := o.ExchangeCode(context.Background(), "code")
	require.Error(t, err)
}

func TestOAuth_ExchangeCode_invalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{`))
	}))
	defer srv.Close()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	o.httpClient = srv.Client()
	o.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		p := strings.TrimPrefix(srv.URL, "http://")
		req.URL.Scheme = "http"
		req.URL.Host = p
		return http.DefaultTransport.RoundTrip(req)
	})

	_, err := o.ExchangeCode(context.Background(), "code")
	require.Error(t, err)
}
