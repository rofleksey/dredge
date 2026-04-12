package httptransport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
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

// LoginRateLimitMiddleware limits POST /auth/login using the client IP (X-Forwarded-For first hop when set).
func LoginRateLimitMiddleware(limiter *LoginLimiter) middleware.Middleware {
	return func(req middleware.Request, next middleware.Next) (middleware.Response, error) {
		if req.OperationName != gen.LoginOperation || limiter == nil {
			return next(req)
		}

		if req.Raw == nil {
			return next(req)
		}

		if !limiter.Allow(clientIP(req.Raw)) {
			return middleware.Response{}, ErrLoginRateLimited
		}

		return next(req)
	}
}

func clientIP(r *http.Request) string {
	// Intentional: first X-Forwarded-For hop is trusted when present. Use only behind a reverse
	// proxy that sets or overwrites X-Forwarded-For from the real client connection.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// OgenErrorHandler wraps the ogen default handler to map known sentinel errors to HTTP statuses.
func OgenErrorHandler() gen.ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, ErrLoginRateLimited) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many login attempts"}`))

			return
		}

		ogenerrors.DefaultErrorHandler(ctx, w, r, err)
	}
}
