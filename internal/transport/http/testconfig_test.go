package httptransport

import "github.com/rofleksey/dredge/internal/config"

func testTwitchServiceConfig(clientID, clientSecret string) config.Config {
	var c config.Config

	c.Twitch.ClientID = clientID
	c.Twitch.ClientSecret = clientSecret
	c.Twitch.OAuthRedirectURI = "http://localhost:8080/oauth/twitch/callback"
	c.Twitch.OAuthReturnURL = "http://localhost:5173/#/settings"
	return c
}
