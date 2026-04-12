package config

import (
	"fmt"
	"net/url"
)

// AllowedOrigin returns the Origin header value (scheme://host[:port]) for a configured base URL.
func AllowedOrigin(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("base_url must include scheme and host")
	}

	return u.Scheme + "://" + u.Host, nil
}

// AllowedWebOrigin is the browser Origin header value for CORS/WebSocket checks, parsed once from Server.BaseURL.
type AllowedWebOrigin string

// ParseAllowedWebOrigin derives AllowedWebOrigin from config (same rules as AllowedOrigin).
func ParseAllowedWebOrigin(cfg Config) (AllowedWebOrigin, error) {
	s, err := AllowedOrigin(cfg.Server.BaseURL)
	if err != nil {
		return "", err
	}

	return AllowedWebOrigin(s), nil
}
