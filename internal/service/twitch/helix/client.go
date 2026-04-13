package helix

import (
	"net/http"
	"sync"
	"time"

	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
)

// Client holds Twitch Helix/API state (app access token, HTTP calls, OAuth refresh helpers).
type Client struct {
	Repo         repository.Store
	Obs          *observability.Stack
	HTTPClient   *http.Client
	ClientID     string
	ClientSecret string
	// UserOAuthTokenCacheTTL caps reuse of refreshed user access tokens (default 30m if zero).
	UserOAuthTokenCacheTTL time.Duration

	appTokenMu  sync.Mutex
	appToken    string
	appTokenExp time.Time

	userOAuthMu sync.Mutex
	userOAuth   map[int64]userOAuthCacheEntry
}

type refreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type userOAuthCacheEntry struct {
	accessToken     string
	refreshSnapshot string
	expiresAt       time.Time
}

// NewClient builds a Helix client with default HTTP timeouts.
func NewClient(repo repository.Store, obs *observability.Stack, clientID, clientSecret string) *Client {
	return &Client{
		Repo:         repo,
		Obs:          obs,
		HTTPClient:   &http.Client{Timeout: 45 * time.Second},
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}
