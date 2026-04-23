package webui

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/rofleksey/dredge/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMux_routesAPIAndStatic(t *testing.T) {
	t.Parallel()

	apiHits := 0

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiHits++

		w.WriteHeader(http.StatusNoContent)
	})

	mux := NewMux(api)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, 1, apiHits)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec2, req2)
	assert.NotEqual(t, http.StatusNotFound, rec2.Code)
}

func TestIsAPIPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{"/api/v1/auth/login", true},
		{"/api/v1/auth/x", true},
		{"/api/v1/me", true},
		{"/api/v1/me/foo", true},
		{"/api/v1/settings/x", true},
		{"/api/v1/twitch/foo", true},
		{"/api/v1/ai/settings", true},
		{"/api/v1/ai/conversations", true},
		{"/", false},
		{"/auth/login", false},
		{"/me", false},
		{"/index.html", false},
		{"/assets/app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, api.IsAPIPath(tt.path))
		})
	}
}

func TestSpaFS_Open_existingFile(t *testing.T) {
	t.Parallel()

	root := fstest.MapFS{
		"app.js":     &fstest.MapFile{Data: []byte("// ok")},
		"index.html": &fstest.MapFile{Data: []byte("<html/>")},
	}

	s := spaFS{root: root}

	f, err := s.Open("app.js")
	require.NoError(t, err)

	require.NoError(t, f.Close())
}

func TestSpaFS_Open_missingWithExtension_returnsNotExist(t *testing.T) {
	t.Parallel()

	root := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html/>")},
	}

	s := spaFS{root: root}

	_, err := s.Open("missing.js")
	require.Error(t, err)
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestSpaFS_Open_missingNoExtension_fallsBackToIndex(t *testing.T) {
	t.Parallel()

	root := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html/>")},
	}

	s := spaFS{root: root}

	f, err := s.Open("spa/route")
	require.NoError(t, err)

	require.NoError(t, f.Close())
}
