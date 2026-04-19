package twitch

import (
	"net/http"

	"github.com/rofleksey/dredge/internal/config"
)

type stopNoopBC struct{}

func (stopNoopBC) BroadcastJSON(any) {}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testTwitchCfg(clientID, clientSecret string) config.Config {
	var c config.Config

	c.Twitch.ClientID = clientID
	c.Twitch.ClientSecret = clientSecret
	c.Twitch.OAuthRedirectURI = "http://localhost:8080/oauth/twitch/callback"
	c.Twitch.OAuthReturnURL = "http://localhost:5173/#/settings"
	return c
}
