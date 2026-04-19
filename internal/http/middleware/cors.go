package httpmw

import "net/http"

// WrapCORS wraps h with CORS headers for a single browser origin (e.g. from config server.base_url).
func WrapCORS(allowOrigin string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := w.Header()
		hdr.Set("Access-Control-Allow-Origin", allowOrigin)
		hdr.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")

		// Intentional: reflect requested headers (or "*") for a single trusted SPA origin + bearer API.
		if reqHdrs := r.Header.Get("Access-Control-Request-Headers"); reqHdrs != "" {
			hdr.Set("Access-Control-Allow-Headers", reqHdrs)
		} else {
			hdr.Set("Access-Control-Allow-Headers", "*")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}
