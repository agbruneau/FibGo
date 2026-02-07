package cli

import (
	"bytes"
	"testing"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/orchestration"
)

// TestPrintExecutionConfig tests the PrintExecutionConfig function.
func TestPrintExecutionConfig(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	cfg := config.AppConfig{
		N:            1000,
		Timeout:      60000000000, // 1 minute
		Threshold:    4096,
		FFTThreshold: 1000000,
	}

	PrintExecutionConfig(cfg, &buf)

	output := buf.String()

	// Check that output contains expected components
	if output == "" {
		t.Error("PrintExecutionConfig should produce output")
	}
	if len(output) < 50 {
		t.Errorf("PrintExecutionConfig output seems too short: %s", output)
	}
}

// TestPrintExecutionMode tests the PrintExecutionMode function.
func TestPrintExecutionMode(t *testing.T) {
	t.Parallel()
	factory := fibonacci.GlobalFactory()

	t.Run("Single calculator mode", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		calculators := []fibonacci.Calculator{factory.MustGet("fast")}

		PrintExecutionMode(calculators, &buf)

		output := buf.String()
		if output == "" {
			t.Error("PrintExecutionMode should produce output")
		}
	})

	t.Run("Multiple calculators mode", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		cfg := config.AppConfig{Algo: "all"}
		calculators := orchestration.GetCalculatorsToRun(cfg, factory)

		PrintExecutionMode(calculators, &buf)

		output := buf.String()
		if output == "" {
			t.Error("PrintExecutionMode should produce output for multiple calculators")
		}
	})
}
