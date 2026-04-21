package httpmw

import "net/http"

// WrapSecurityHeaders sets baseline HTTP response headers for the API and embedded SPA.
func WrapSecurityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := w.Header()
		hdr.Set("X-Content-Type-Options", "nosniff")
		hdr.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		hdr.Set("X-Frame-Options", "DENY")
		// img-src / frame-src: Twitch avatars and embedded player (Helix URLs + player.twitch.tv).
		hdr.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; "+
			"script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data: https:; "+
			"frame-src https://player.twitch.tv; worker-src 'self' blob:")

		h.ServeHTTP(w, r)
	})
}
