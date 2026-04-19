package twitch

import (
	"net/url"
	"strings"
)

// AuthorizeURL is the Twitch page where the user approves scopes (twitchium-style authorize URL).
func (o *OAuth) AuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", o.clientID)
	params.Set("redirect_uri", o.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(o.scopes, " "))
	params.Set("state", state)

	return "https://id.twitch.tv/oauth2/authorize?" + params.Encode()
}
