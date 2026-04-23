package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewMux serves the embedded SPA. Register the OpenAPI handler on "/api/v1/" in the root mux.
func NewMux() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("webui: static embed: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(spaFS{root: sub}))
	return fileServer
}

// spaFS wraps fs.FS to serve index.html for missing paths (Vue router).
type spaFS struct {
	root fs.FS
}

func (s spaFS) Open(name string) (fs.File, error) {
	f, err := s.root.Open(name)
	if err == nil {
		return f, nil
	}

	base := path.Base(name)
	if strings.Contains(base, ".") && base != "index.html" {
		return nil, err
	}
	return s.root.Open("index.html")
}
