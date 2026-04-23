package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAPIPath(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, PrefixAI)
	assert.NotEmpty(t, PrefixAuth)
	assert.NotEmpty(t, PrefixSettings)
	assert.NotEmpty(t, PrefixTwitch)

	cases := []struct {
		path string
		want bool
	}{
		{"/api/v1/auth/login", true},
		{"/api/v1/me", true},
		{"/api/v1/me/x", true},
		{"/api/v1/settings/rules", true},
		{"/api/v1/twitch/users", true},
		{"/api/v1/ai/settings", true},
		{"/api/v1/ai/conversations", true},
		{"/api/v1/ai/conversations/1/messages", true},
		{"/", false},
		{"/auth/login", false},
		{"/me", false},
		{"/assets/foo.js", false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, IsAPIPath(tc.path), "path %q", tc.path)
	}
}
