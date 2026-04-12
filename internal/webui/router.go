package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	httproutes "github.com/rofleksey/dredge/internal/api"
)

// NewMux returns a mux that serves the OpenAPI API and the embedded SPA.
func NewMux(apiHandler http.Handler) http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("webui: static embed: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(spaFS{root: sub}))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httproutes.IsAPIPath(r.URL.Path) {
			apiHandler.ServeHTTP(w, r)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
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
