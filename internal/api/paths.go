package api

import "strings"

// API route prefixes (see api/openapi.yaml). Used by the SPA mux to route to ogen.
const (
	PrefixAuth     = "/auth/"
	PrefixMe       = "/me"
	PrefixSettings = "/settings/"
	PrefixTwitch   = "/twitch/"
)

// IsAPIPath reports whether the request path is served by the OpenAPI HTTP API (not static files).
func IsAPIPath(p string) bool {
	switch {
	case strings.HasPrefix(p, PrefixAuth):
		return true
	case p == PrefixMe || strings.HasPrefix(p, PrefixMe+"/"):
		return true
	case strings.HasPrefix(p, PrefixSettings):
		return true
	case strings.HasPrefix(p, PrefixTwitch):
		return true
	default:
		return false
	}
}
