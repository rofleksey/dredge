package httptransport

import (
	"net/http"
	"net/http/httptest"
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

func TestClientIP_xForwardedFor(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	assert.Equal(t, "203.0.113.1", clientIP(req))
}
