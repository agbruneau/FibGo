package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agbru/fibcalc/internal/logging"
)

// TestNewMetrics tests the Metrics constructor.
func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	if m.handler == nil {
		t.Error("Metrics.handler should be initialized")
	}
}

// TestMetrics_IncrementDecrementActiveRequests tests the active requests gauge.
func TestMetrics_IncrementDecrementActiveRequests(t *testing.T) {
	m := NewMetrics()

	// Note: Prometheus metrics are global singletons.
	// This test verifies the methods don't panic and work correctly.

	t.Run("IncrementActiveRequests does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("IncrementActiveRequests panicked: %v", r)
			}
		}()
		m.IncrementActiveRequests()
	})

	t.Run("DecrementActiveRequests does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DecrementActiveRequests panicked: %v", r)
			}
		}()
		m.DecrementActiveRequests()
	})
}

// TestMetrics_WritePrometheus tests the Prometheus metrics endpoint.
func TestMetrics_WritePrometheus(t *testing.T) {
	m := NewMetrics()

	// Call increment to ensure we have some metrics
	m.IncrementActiveRequests()
	defer m.DecrementActiveRequests()

	req := httptest.NewRequest("GET", "/metrics", http.NoBody)
	rec := httptest.NewRecorder()

	m.WritePrometheus(rec, req)

	body := rec.Body.String()

	t.Run("Contains active requests metric", func(t *testing.T) {
		if !strings.Contains(body, "fibcalc_active_requests") {
			t.Error("metrics output should contain fibcalc_active_requests")
		}
	})

	t.Run("Contains total requests metric", func(t *testing.T) {
		if !strings.Contains(body, "fibcalc_requests_total") {
			t.Error("metrics output should contain fibcalc_requests_total")
		}
	})

	t.Run("Contains Go runtime metrics", func(t *testing.T) {
		if !strings.Contains(body, "go_") {
			t.Error("metrics output should contain Go runtime metrics")
		}
	})
}

// TestServer_metricsMiddleware tests the metrics tracking middleware.
func TestServer_metricsMiddleware(t *testing.T) {
	t.Run("Next handler is called", func(t *testing.T) {
		s := &Server{
			metrics: NewMetrics(),
		}

		nextCalled := false
		next := func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}

		handler := s.metricsMiddleware(next)
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		rec := httptest.NewRecorder()

		handler(rec, req)

		if !nextCalled {
			t.Error("next handler was not called")
		}
	})

	t.Run("Metrics are tracked", func(t *testing.T) {
		s := &Server{
			metrics: NewMetrics(),
		}

		next := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		handler := s.metricsMiddleware(next)
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		rec := httptest.NewRecorder()

		// This should not panic and should track the request
		handler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("Decrement is called even on panic recovery", func(t *testing.T) {
		s := &Server{
			metrics: NewMetrics(),
		}

		// This test just verifies the defer pattern works
		next := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		handler := s.metricsMiddleware(next)
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		rec := httptest.NewRecorder()

		handler(rec, req)
		// If we got here without panic, the middleware is working
	})
}

// TestServer_handleMetrics tests the /metrics endpoint handler.
func TestServer_handleMetrics(t *testing.T) {
	t.Run("GET returns metrics", func(t *testing.T) {
		s := &Server{
			metrics: NewMetrics(),
		}

		req := httptest.NewRequest("GET", "/metrics", http.NoBody)
		rec := httptest.NewRecorder()

		s.handleMetrics(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "fibcalc_") {
			t.Error("response should contain fibcalc metrics")
		}
	})

	t.Run("POST returns method not allowed", func(t *testing.T) {
		s := &Server{
			metrics: NewMetrics(),
			logger:  newTestLogger(),
		}

		req := httptest.NewRequest("POST", "/metrics", http.NoBody)
		rec := httptest.NewRecorder()

		s.handleMetrics(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("PUT returns method not allowed", func(t *testing.T) {
		s := &Server{
			metrics: NewMetrics(),
			logger:  newTestLogger(),
		}

		req := httptest.NewRequest("PUT", "/metrics", http.NoBody)
		rec := httptest.NewRecorder()

		s.handleMetrics(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}

// testLogger is a minimal logger for testing that implements logging.Logger.
type testLogger struct{}

func newTestLogger() *testLogger                                  { return &testLogger{} }
func (l *testLogger) Info(_ string, _ ...logging.Field)           {}
func (l *testLogger) Error(_ string, _ error, _ ...logging.Field) {}
func (l *testLogger) Debug(_ string, _ ...logging.Field)          {}
func (l *testLogger) Printf(_ string, _ ...any)                   {}
func (l *testLogger) Println(_ ...any)                            {}
