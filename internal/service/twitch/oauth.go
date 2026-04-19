package twitch

import (
	"net/http"
	"sync"
	"time"
)

// OAuth performs the Twitch authorization-code flow (same token endpoints as twitchium).
type OAuth struct {
	clientID     string
	clientSecret string
	redirectURI  string
	returnURL    string
	scopes       []string
	hmacSecret   string
	httpClient   *http.Client

	stateMu sync.Mutex
	// Intentional: single-process nonce store; multiple app instances do not share replay state.
	usedStateNonces map[string]int64 // nonce -> state expiry (unix); single-use within TTL
}

// NewOAuth builds the OAuth helper; twitch.oauth_redirect_uri and twitch.oauth_return_url must be set in config (validated at load).
func NewOAuth(clientID, clientSecret, redirectURI, returnURL, stateHMACSecret string) *OAuth {
	return &OAuth{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		returnURL:    returnURL,
		scopes:       UserOAuthScopes,
		hmacSecret:   stateHMACSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}
