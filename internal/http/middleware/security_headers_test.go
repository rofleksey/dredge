package httpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapSecurityHeaders_setsExpectedHeaders(t *testing.T) {
	t.Parallel()

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	srv := httptest.NewServer(WrapSecurityHeaders(inner))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusTeapot, resp.StatusCode)
	require.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	require.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	require.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	csp := resp.Header.Get("Content-Security-Policy")
	require.Contains(t, csp, "default-src 'self'")
	require.Contains(t, csp, "frame-ancestors 'none'")
	require.Contains(t, csp, "https://player.twitch.tv")
}
