package httpmw

import (
	"errors"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

const loginLimiterMaxIPs = 10_000

// ErrLoginRateLimited is returned by login rate-limit middleware before the handler runs.
var ErrLoginRateLimited = errors.New("login rate limited")

// LoginLimiter enforces a rolling per-minute cap per client IP for login attempts.
type LoginLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	byIP   *lru.Cache[string, []time.Time]
}

// NewLoginLimiter builds a limiter; limit 0 disables checking (always allow).
func NewLoginLimiter(loginPerMinute int) *LoginLimiter {
	cache, err := lru.New[string, []time.Time](loginLimiterMaxIPs)
	if err != nil {
		panic(err)
	}

	return &LoginLimiter{
		limit:  loginPerMinute,
		window: time.Minute,
		byIP:   cache,
	}
}

// Allow records an attempt and reports whether it is within the limit.
func (l *LoginLimiter) Allow(ip string) bool {
	if l == nil || l.limit <= 0 {
		return true
	}

	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	var times []time.Time
	if v, ok := l.byIP.Get(ip); ok {
		times = v
	}

	cutoff := now.Add(-l.window)

	kept := make([]time.Time, 0, len(times)+1)
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= l.limit {
		l.byIP.Add(ip, kept)
		return false
	}

	kept = append(kept, now)
	l.byIP.Add(ip, kept)

	return true
}
