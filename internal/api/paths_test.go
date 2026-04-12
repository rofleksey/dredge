package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAPIPath(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, PrefixAuth)
	assert.NotEmpty(t, PrefixSettings)
	assert.NotEmpty(t, PrefixTwitch)

	cases := []struct {
		path string
		want bool
	}{
		{"/auth/login", true},
		{"/me", true},
		{"/me/x", true},
		{"/settings/rules", true},
		{"/twitch/users", true},
		{"/", false},
		{"/assets/foo.js", false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, IsAPIPath(tc.path), "path %q", tc.path)
	}
}
