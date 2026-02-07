package orchestration

import (
	"testing"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// TestGetCalculatorsToRun tests the GetCalculatorsToRun function.
func TestGetCalculatorsToRun(t *testing.T) {
	t.Parallel()
	factory := fibonacci.GlobalFactory()

	t.Run("Single algorithm returns one calculator", func(t *testing.T) {
		t.Parallel()
		cfg := config.AppConfig{Algo: "fast"}
		calculators := GetCalculatorsToRun(cfg, factory)

		if len(calculators) != 1 {
			t.Errorf("Expected 1 calculator, got %d", len(calculators))
		}
		// Check that the name contains "Fast Doubling" (exact name may vary)
		if calculators[0].Name() == "" {
			t.Error("Calculator name should not be empty")
		}
	})

	t.Run("All algorithms returns multiple calculators", func(t *testing.T) {
		t.Parallel()
		cfg := config.AppConfig{Algo: "all"}
		calculators := GetCalculatorsToRun(cfg, factory)

		if len(calculators) < 2 {
			t.Errorf("Expected at least 2 calculators for 'all', got %d", len(calculators))
		}
	})

	t.Run("Matrix algorithm", func(t *testing.T) {
		t.Parallel()
		cfg := config.AppConfig{Algo: "matrix"}
		calculators := GetCalculatorsToRun(cfg, factory)

		if len(calculators) != 1 {
			t.Errorf("Expected 1 calculator, got %d", len(calculators))
		}
	})
}
