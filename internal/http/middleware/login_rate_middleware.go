package httpmw

import (
	"net"
	"net/http"
	"strings"

	ogenmw "github.com/ogen-go/ogen/middleware"

	"github.com/rofleksey/dredge/internal/http/gen"
)

// LoginRateLimitMiddleware limits POST /auth/login using the client IP (X-Forwarded-For first hop when set).
func LoginRateLimitMiddleware(limiter *LoginLimiter) ogenmw.Middleware {
	return func(req ogenmw.Request, next ogenmw.Next) (ogenmw.Response, error) {
		if req.OperationName != gen.LoginOperation || limiter == nil {
			return next(req)
		}

		if req.Raw == nil {
			return next(req)
		}

		if !limiter.Allow(clientIP(req.Raw)) {
			return ogenmw.Response{}, ErrLoginRateLimited
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
