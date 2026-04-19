package httpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapCORS_options(t *testing.T) {
	t.Parallel()

	const origin = "http://localhost:5173"

	h := WrapCORS(origin, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next must not run for OPTIONS")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestWrapCORS_get(t *testing.T) {
	t.Parallel()

	const origin = "http://localhost:5173"

	h := WrapCORS(origin, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))
}
