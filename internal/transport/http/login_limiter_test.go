package httptransport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginLimiter_disabled(t *testing.T) {
	t.Parallel()

	l := NewLoginLimiter(0)
	for range 20 {
		assert.True(t, l.Allow("1.2.3.4"))
	}
}

func TestLoginLimiter_perMinute(t *testing.T) {
	t.Parallel()

	l := NewLoginLimiter(3)
	ip := "10.0.0.1"

	assert.True(t, l.Allow(ip))
	assert.True(t, l.Allow(ip))
	assert.True(t, l.Allow(ip))
	assert.False(t, l.Allow(ip))
}
