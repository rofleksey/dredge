package observability

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func (s *Stack) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func (s *Stack) InstrumentHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		path := r.URL.Path
		if path == "" {
			path = "/"
		}

		ww := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(ww, r)

		code := ww.status
		if code == 0 {
			code = http.StatusOK
		}

		status := fmt.Sprintf("%d", code)
		s.requestsTotal.WithLabelValues(r.Method, path, status).Inc()
		s.requestLatency.WithLabelValues(r.Method, path, status).Observe(time.Since(start).Seconds())
		s.Logger.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", code),
			zap.Duration("latency", time.Since(start)),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}

	r.wroteHeader = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

// Hijack implements [http.Hijacker] so WebSocket upgrades work through this wrapper.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}
