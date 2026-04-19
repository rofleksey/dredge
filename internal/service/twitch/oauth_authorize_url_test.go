package twitch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth_AuthorizeURL(t *testing.T) {
	t.Parallel()

	o := NewOAuth("myid", "sec", "http://localhost/oauth/callback", "http://localhost/#/x", "sixteen-byte-key!!")

	u := o.AuthorizeURL("st")
	assert.Contains(t, u, "https://id.twitch.tv/oauth2/authorize")
	assert.Contains(t, u, "client_id=myid")
	assert.Contains(t, u, "state=st")
}
