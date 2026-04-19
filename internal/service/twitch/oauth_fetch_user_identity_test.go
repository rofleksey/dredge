package twitch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth_FetchUserIdentity(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"12345","login":"HelloWorld"}]}`))
	}))
	defer srv.Close()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")
	o.httpClient = srv.Client()

	o.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	})

	id, login, err := o.FetchUserIdentity(context.Background(), "token")
	require.NoError(t, err)
	assert.Equal(t, int64(12345), id)
	assert.Equal(t, "helloworld", login)
}

func TestOAuth_FetchUserIdentity_emptyData(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
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

	_, _, err := o.FetchUserIdentity(context.Background(), "tok")
	require.Error(t, err)
}

func TestOAuth_FetchUserIdentity_badStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
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

	_, _, err := o.FetchUserIdentity(context.Background(), "tok")
	require.Error(t, err)
}
