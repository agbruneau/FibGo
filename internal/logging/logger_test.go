package logging

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// TestFieldHelpers tests the Field constructor functions.
func TestFieldHelpers(t *testing.T) {
	t.Run("String creates field with key and string value", func(t *testing.T) {
		f := String("key", "value")
		if f.Key != "key" {
			t.Errorf("String().Key = %q, want %q", f.Key, "key")
		}
		if f.Value != "value" {
			t.Errorf("String().Value = %q, want %q", f.Value, "value")
		}
	})

	t.Run("Int creates field with key and int value", func(t *testing.T) {
		f := Int("count", 42)
		if f.Key != "count" {
			t.Errorf("Int().Key = %q, want %q", f.Key, "count")
		}
		if f.Value != 42 {
			t.Errorf("Int().Value = %v, want %v", f.Value, 42)
		}
	})

	t.Run("Uint64 creates field with key and uint64 value", func(t *testing.T) {
		f := Uint64("n", 12345678901234567890)
		if f.Key != "n" {
			t.Errorf("Uint64().Key = %q, want %q", f.Key, "n")
		}
		if f.Value != uint64(12345678901234567890) {
			t.Errorf("Uint64().Value = %v, want %v", f.Value, uint64(12345678901234567890))
		}
	})

	t.Run("Float64 creates field with key and float64 value", func(t *testing.T) {
		f := Float64("duration", 3.14159)
		if f.Key != "duration" {
			t.Errorf("Float64().Key = %q, want %q", f.Key, "duration")
		}
		if f.Value != 3.14159 {
			t.Errorf("Float64().Value = %v, want %v", f.Value, 3.14159)
		}
	})

	t.Run("Err creates field with error key", func(t *testing.T) {
		testErr := errors.New("test error")
		f := Err(testErr)
		if f.Key != "error" {
			t.Errorf("Err().Key = %q, want %q", f.Key, "error")
		}
		if f.Value != testErr {
			t.Errorf("Err().Value = %v, want %v", f.Value, testErr)
		}
	})

	t.Run("Err with nil error", func(t *testing.T) {
		f := Err(nil)
		if f.Key != "error" {
			t.Errorf("Err(nil).Key = %q, want %q", f.Key, "error")
		}
		if f.Value != nil {
			t.Errorf("Err(nil).Value = %v, want nil", f.Value)
		}
	})
}

// TestNewZerologAdapter tests the ZerologAdapter constructor.
func TestNewZerologAdapter(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf)
	adapter := NewZerologAdapter(zl)

	if adapter == nil {
		t.Fatal("NewZerologAdapter returned nil")
	}

	adapter.Info("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("NewZerologAdapter logger not working, output: %s", buf.String())
	}
}

// TestNewDefaultLogger tests the default logger constructor.
func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	if logger == nil {
		t.Fatal("NewDefaultLogger returned nil")
	}
}

// TestNewLogger tests the custom logger constructor.
func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "test-component")

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	logger.Info("hello")
	output := buf.String()

	if !strings.Contains(output, "test-component") {
		t.Errorf("NewLogger should include component field, got: %s", output)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("NewLogger should include message, got: %s", output)
	}
}

// TestZerologAdapter_Info tests the Info method.
func TestZerologAdapter_Info(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		fields   []Field
		contains []string
	}{
		{
			name:     "no fields",
			msg:      "test message",
			fields:   nil,
			contains: []string{"test message", "info"},
		},
		{
			name:     "with string field",
			msg:      "user login",
			fields:   []Field{String("user", "alice")},
			contains: []string{"user login", "alice"},
		},
		{
			name:     "with multiple fields",
			msg:      "request processed",
			fields:   []Field{String("method", "GET"), Int("status", 200)},
			contains: []string{"request processed", "GET", "200"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, "test")
			logger.Info(tt.msg, tt.fields...)

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestZerologAdapter_Error tests the Error method.
func TestZerologAdapter_Error(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		err      error
		fields   []Field
		contains []string
	}{
		{
			name:     "with error",
			msg:      "operation failed",
			err:      errors.New("connection refused"),
			fields:   nil,
			contains: []string{"operation failed", "connection refused", "error"},
		},
		{
			name:     "with nil error",
			msg:      "warning",
			err:      nil,
			fields:   nil,
			contains: []string{"warning", "error"},
		},
		{
			name:     "with error and fields",
			msg:      "db error",
			err:      errors.New("timeout"),
			fields:   []Field{String("db", "postgres"), Int("retry", 3)},
			contains: []string{"db error", "timeout", "postgres", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, "test")
			logger.Error(tt.msg, tt.err, tt.fields...)

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestZerologAdapter_Debug tests the Debug method.
func TestZerologAdapter_Debug(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel)
	logger := NewZerologAdapter(zl)

	logger.Debug("debug message", String("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Debug output should contain message, got: %s", output)
	}
	if !strings.Contains(output, "debug") {
		t.Errorf("Debug output should contain level, got: %s", output)
	}
}

// TestZerologAdapter_Printf tests the Printf method.
func TestZerologAdapter_Printf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "test")

	logger.Printf("formatted %s %d", "message", 42)

	output := buf.String()
	if !strings.Contains(output, "formatted message 42") {
		t.Errorf("Printf should format message, got: %s", output)
	}
}

// TestZerologAdapter_Println tests the Println method.
func TestZerologAdapter_Println(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "test")

	logger.Println("hello", "world")

	output := buf.String()
	if !strings.Contains(output, "hello") || !strings.Contains(output, "world") {
		t.Errorf("Println should include all arguments, got: %s", output)
	}
}

// TestZerologAdapter_applyFields tests field application with all supported types.
func TestZerologAdapter_applyFields(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		contains string
	}{
		{"string field", Field{Key: "str", Value: "hello"}, "hello"},
		{"int field", Field{Key: "num", Value: 42}, "42"},
		{"int64 field", Field{Key: "big", Value: int64(9223372036854775807)}, "9223372036854775807"},
		{"uint64 field", Field{Key: "huge", Value: uint64(18446744073709551615)}, "18446744073709551615"},
		{"float64 field", Field{Key: "pi", Value: 3.14}, "3.14"},
		{"error field", Field{Key: "err", Value: errors.New("oops")}, "oops"},
		{"bool field", Field{Key: "flag", Value: true}, "true"},
		{"interface field", Field{Key: "data", Value: struct{ X int }{X: 1}}, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, "test")
			logger.Info("test", tt.field)

			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("applyFields should handle %s, output: %s", tt.name, output)
			}
		})
	}
}

// TestNewStdLoggerAdapter tests the StdLoggerAdapter constructor.
func TestNewStdLoggerAdapter(t *testing.T) {
	var buf bytes.Buffer
	stdLogger := log.New(&buf, "", 0)
	adapter := NewStdLoggerAdapter(stdLogger)

	if adapter == nil {
		t.Fatal("NewStdLoggerAdapter returned nil")
	}

	adapter.Info("test")
	if !strings.Contains(buf.String(), "test") {
		t.Errorf("StdLoggerAdapter not working, output: %s", buf.String())
	}
}

// TestStdLoggerAdapter_Info tests the StdLoggerAdapter Info method.
func TestStdLoggerAdapter_Info(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		fields   []Field
		contains []string
	}{
		{
			name:     "no fields",
			msg:      "info message",
			fields:   nil,
			contains: []string{"[INFO]", "info message"},
		},
		{
			name:     "with fields",
			msg:      "user action",
			fields:   []Field{String("user", "bob")},
			contains: []string{"[INFO]", "user action", "user", "bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			stdLogger := log.New(&buf, "", 0)
			adapter := NewStdLoggerAdapter(stdLogger)

			adapter.Info(tt.msg, tt.fields...)

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestStdLoggerAdapter_Error tests the StdLoggerAdapter Error method.
func TestStdLoggerAdapter_Error(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		err      error
		fields   []Field
		contains []string
	}{
		{
			name:     "with error no fields",
			msg:      "failed",
			err:      errors.New("boom"),
			fields:   nil,
			contains: []string{"[ERROR]", "failed", "boom"},
		},
		{
			name:     "with error and fields",
			msg:      "db failed",
			err:      errors.New("timeout"),
			fields:   []Field{String("db", "mysql")},
			contains: []string{"[ERROR]", "db failed", "timeout", "mysql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			stdLogger := log.New(&buf, "", 0)
			adapter := NewStdLoggerAdapter(stdLogger)

			adapter.Error(tt.msg, tt.err, tt.fields...)

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestStdLoggerAdapter_Debug tests the StdLoggerAdapter Debug method.
func TestStdLoggerAdapter_Debug(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		fields   []Field
		contains []string
	}{
		{
			name:     "no fields",
			msg:      "debug info",
			fields:   nil,
			contains: []string{"[DEBUG]", "debug info"},
		},
		{
			name:     "with fields",
			msg:      "trace",
			fields:   []Field{Int("line", 42)},
			contains: []string{"[DEBUG]", "trace", "line", "42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			stdLogger := log.New(&buf, "", 0)
			adapter := NewStdLoggerAdapter(stdLogger)

			adapter.Debug(tt.msg, tt.fields...)

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// TestStdLoggerAdapter_Printf tests the StdLoggerAdapter Printf method.
func TestStdLoggerAdapter_Printf(t *testing.T) {
	var buf bytes.Buffer
	stdLogger := log.New(&buf, "", 0)
	adapter := NewStdLoggerAdapter(stdLogger)

	adapter.Printf("value is %d", 123)

	output := buf.String()
	if !strings.Contains(output, "value is 123") {
		t.Errorf("Printf should format string, got: %s", output)
	}
}

// TestStdLoggerAdapter_Println tests the StdLoggerAdapter Println method.
func TestStdLoggerAdapter_Println(t *testing.T) {
	var buf bytes.Buffer
	stdLogger := log.New(&buf, "", 0)
	adapter := NewStdLoggerAdapter(stdLogger)

	adapter.Println("a", "b", "c")

	output := buf.String()
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") || !strings.Contains(output, "c") {
		t.Errorf("Println should include all args, got: %s", output)
	}
}

// TestLoggerInterface verifies both adapters implement the Logger interface.
func TestLoggerInterface(t *testing.T) {
	t.Run("ZerologAdapter implements Logger", func(t *testing.T) {
		var buf bytes.Buffer
		var _ Logger = NewLogger(&buf, "test")
	})

	t.Run("StdLoggerAdapter implements Logger", func(t *testing.T) {
		var buf bytes.Buffer
		stdLogger := log.New(&buf, "", 0)
		var _ Logger = NewStdLoggerAdapter(stdLogger)
	})
}
