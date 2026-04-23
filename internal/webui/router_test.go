package webui

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMux_servesStatic(t *testing.T) {
	t.Parallel()

	mux := NewMux()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(rec, req)
	assert.NotEqual(t, http.StatusNotFound, rec.Code)
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
