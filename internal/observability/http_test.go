package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func TestStack_MetricsHandler(t *testing.T) {
	t.Parallel()

	s := testStack(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	s.MetricsHandler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStack_InstrumentHTTP(t *testing.T) {
	t.Parallel()

	s := testStack(t)
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	h := s.InstrumentHTTP(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTeapot, rec.Code)
}

func TestStack_InstrumentHTTP_writeWithoutHeader(t *testing.T) {
	t.Parallel()

	s := testStack(t)
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	h := s.InstrumentHTTP(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/z", nil)

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func testStack(t *testing.T) *Stack {
	t.Helper()

	reg := prometheus.NewRegistry()
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "dredge_http_requests_total_test", Help: "test"},
		[]string{"method", "path", "status"},
	)
	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "dredge_http_request_duration_seconds_test", Help: "test"},
		[]string{"method", "path", "status"},
	)
	reg.MustRegister(requestsTotal, requestLatency)

	return &Stack{
		Logger:         zap.NewNop(),
		Tracer:         otel.Tracer("test"),
		requestsTotal:  requestsTotal,
		requestLatency: requestLatency,
	}
}
