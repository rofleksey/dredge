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

func TestOAuth_NewState_VerifyState(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	state, err := o.NewState("")
	require.NoError(t, err)
	require.NotEmpty(t, state)

	_, err = o.VerifyState(state)
	require.NoError(t, err)
}

func TestOAuth_VerifyState_rejectsReuse(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	state, err := o.NewState("")
	require.NoError(t, err)

	_, err = o.VerifyState(state)
	require.NoError(t, err)
	_, err = o.VerifyState(state)
	assert.Error(t, err)
}

func TestOAuth_VerifyState_errors(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	_, err := o.VerifyState("bad")
	assert.Error(t, err)
	_, err = o.VerifyState("a.b.c")
	assert.Error(t, err)
}

func TestOAuth_VerifyState_signatureMismatch(t *testing.T) {
	t.Parallel()

	o := NewOAuth("cid", "sec", "http://localhost/cb", "http://localhost/app", "sixteen-byte-key!!")

	st, err := o.NewState("")
	require.NoError(t, err)

	parts := strings.Split(st, ".")
	require.Len(t, parts, 2)

	tampered := parts[0] + ".YmFk" // "bad" base64 raw

	_, err = o.VerifyState(tampered)
	assert.Error(t, err)
}

func TestOAuth_AuthorizeURL(t *testing.T) {
	t.Parallel()

	o := NewOAuth("myid", "sec", "http://localhost/oauth/callback", "http://localhost/#/x", "sixteen-byte-key!!")

	u := o.AuthorizeURL("st")
	assert.Contains(t, u, "https://id.twitch.tv/oauth2/authorize")
	assert.Contains(t, u, "client_id=myid")
	assert.Contains(t, u, "state=st")
}

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

func TestOAuth_ReturnURL(t *testing.T) {
	t.Parallel()

	o := NewOAuth("c", "s", "http://a", "http://return/here", "sixteen-byte-key!!")
	assert.Equal(t, "http://return/here", o.ReturnURL())
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
