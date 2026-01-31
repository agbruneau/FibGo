package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDefaultSecurityConfig verifies default security configuration values.
func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	t.Run("EnableCORS is true", func(t *testing.T) {
		if !config.EnableCORS {
			t.Error("EnableCORS should be true by default")
		}
	})

	t.Run("AllowedOrigins contains wildcard", func(t *testing.T) {
		if len(config.AllowedOrigins) != 1 || config.AllowedOrigins[0] != "*" {
			t.Errorf("AllowedOrigins = %v, want [\"*\"]", config.AllowedOrigins)
		}
	})

	t.Run("AllowedMethods contains GET and OPTIONS", func(t *testing.T) {
		if len(config.AllowedMethods) != 2 {
			t.Errorf("AllowedMethods = %v, want [\"GET\", \"OPTIONS\"]", config.AllowedMethods)
		}
		hasGet := false
		hasOptions := false
		for _, m := range config.AllowedMethods {
			if m == "GET" {
				hasGet = true
			}
			if m == "OPTIONS" {
				hasOptions = true
			}
		}
		if !hasGet || !hasOptions {
			t.Errorf("AllowedMethods = %v, want [\"GET\", \"OPTIONS\"]", config.AllowedMethods)
		}
	})

	t.Run("MaxNValue is 1 billion", func(t *testing.T) {
		if config.MaxNValue != 1_000_000_000 {
			t.Errorf("MaxNValue = %d, want %d", config.MaxNValue, 1_000_000_000)
		}
	})
}

// TestSecurityMiddleware_SecurityHeaders tests that all security headers are set.
func TestSecurityMiddleware_SecurityHeaders(t *testing.T) {
	config := DefaultSecurityConfig()
	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}

	handler := SecurityMiddleware(config, next)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	tests := []struct {
		header string
		want   string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := rec.Header().Get(tt.header)
			if got != tt.want {
				t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}

	if !nextCalled {
		t.Error("next handler was not called")
	}
}

// TestSecurityMiddleware_CORS tests CORS header handling.
func TestSecurityMiddleware_CORS(t *testing.T) {
	tests := []struct {
		name              string
		config            SecurityConfig
		origin            string
		expectCORSHeaders bool
		expectedOrigin    string
	}{
		{
			name: "CORS disabled",
			config: SecurityConfig{
				EnableCORS: false,
			},
			origin:            "http://example.com",
			expectCORSHeaders: false,
		},
		{
			name: "CORS with wildcard origin",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST"},
			},
			origin:            "http://example.com",
			expectCORSHeaders: true,
			expectedOrigin:    "*",
		},
		{
			name: "CORS with specific allowed origin",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"http://allowed.com"},
				AllowedMethods: []string{"GET"},
			},
			origin:            "http://allowed.com",
			expectCORSHeaders: true,
			expectedOrigin:    "http://allowed.com",
		},
		{
			name: "CORS with disallowed origin",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"http://allowed.com"},
				AllowedMethods: []string{"GET"},
			},
			origin:            "http://notallowed.com",
			expectCORSHeaders: false,
		},
		{
			name: "CORS with multiple origins - first match",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"http://first.com", "http://second.com"},
				AllowedMethods: []string{"GET"},
			},
			origin:            "http://first.com",
			expectCORSHeaders: true,
			expectedOrigin:    "http://first.com",
		},
		{
			name: "CORS with multiple origins - second match",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"http://first.com", "http://second.com"},
				AllowedMethods: []string{"GET"},
			},
			origin:            "http://second.com",
			expectCORSHeaders: true,
			expectedOrigin:    "http://second.com",
		},
		{
			name: "CORS with no origin header and wildcard",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET"},
			},
			origin:            "",
			expectCORSHeaders: true, // Wildcard "*" always matches
			expectedOrigin:    "*",
		},
		{
			name: "CORS with no origin header and specific origins",
			config: SecurityConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"http://specific.com"},
				AllowedMethods: []string{"GET"},
			},
			origin:            "",
			expectCORSHeaders: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := func(w http.ResponseWriter, r *http.Request) {}
			handler := SecurityMiddleware(tt.config, next)

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			corsOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.expectCORSHeaders {
				if corsOrigin != tt.expectedOrigin {
					t.Errorf("Access-Control-Allow-Origin = %q, want %q", corsOrigin, tt.expectedOrigin)
				}
				if rec.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("Access-Control-Allow-Methods should be set")
				}
				if rec.Header().Get("Access-Control-Allow-Headers") == "" {
					t.Error("Access-Control-Allow-Headers should be set")
				}
				if rec.Header().Get("Access-Control-Max-Age") == "" {
					t.Error("Access-Control-Max-Age should be set")
				}
			} else if corsOrigin != "" {
				t.Errorf("Access-Control-Allow-Origin should be empty, got %q", corsOrigin)
			}
		})
	}
}

// TestSecurityMiddleware_Preflight tests OPTIONS preflight handling.
func TestSecurityMiddleware_Preflight(t *testing.T) {
	t.Run("OPTIONS returns 204 No Content", func(t *testing.T) {
		config := DefaultSecurityConfig()
		nextCalled := false
		next := func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		}

		handler := SecurityMiddleware(config, next)
		req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if nextCalled {
			t.Error("next handler should not be called for OPTIONS")
		}
	})

	t.Run("OPTIONS sets CORS headers", func(t *testing.T) {
		config := DefaultSecurityConfig()
		next := func(w http.ResponseWriter, r *http.Request) {}

		handler := SecurityMiddleware(config, next)
		req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("CORS headers should be set for OPTIONS")
		}
	})

	t.Run("Non-OPTIONS calls next handler", func(t *testing.T) {
		config := DefaultSecurityConfig()
		nextCalled := false
		next := func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		}

		handler := SecurityMiddleware(config, next)
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		rec := httptest.NewRecorder()

		handler(rec, req)

		if !nextCalled {
			t.Error("next handler should be called for non-OPTIONS")
		}
	})
}

// TestSecurityMiddleware_NextHandlerCalled verifies the next handler is called.
func TestSecurityMiddleware_NextHandlerCalled(t *testing.T) {
	config := DefaultSecurityConfig()
	responseBody := "hello from next"
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}

	handler := SecurityMiddleware(config, next)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != responseBody {
		t.Errorf("body = %q, want %q", rec.Body.String(), responseBody)
	}
}

// TestSecurityMiddleware_AllMethods tests various HTTP methods.
func TestSecurityMiddleware_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			config := DefaultSecurityConfig()
			nextCalled := false
			next := func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			}

			handler := SecurityMiddleware(config, next)
			req := httptest.NewRequest(method, "/test", http.NoBody)
			rec := httptest.NewRecorder()

			handler(rec, req)

			if !nextCalled {
				t.Errorf("next handler should be called for %s", method)
			}
			// Verify security headers are set regardless of method
			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("security headers should be set for %s", method)
			}
		})
	}
}
