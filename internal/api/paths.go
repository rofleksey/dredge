package api

import "strings"

// API route prefixes (see api/openapi.yaml). Used by the SPA mux to route to ogen.
const (
	PrefixAPI      = "/api/v1/"
	PrefixAI       = PrefixAPI + "ai/"
	PrefixAuth     = PrefixAPI + "auth/"
	PrefixMe       = PrefixAPI + "me"
	PrefixSettings = PrefixAPI + "settings/"
	PrefixTwitch   = PrefixAPI + "twitch/"
)

// IsAPIPath reports whether the request path is served by the OpenAPI HTTP API (not static files).
func IsAPIPath(p string) bool {
	return strings.HasPrefix(p, PrefixAPI)
}
