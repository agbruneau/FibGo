package orchestration

import (
	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// GetCalculatorsToRun determines which calculators should be executed based on
// the configuration. Returns calculators in alphabetically sorted order for
// consistent, reproducible behavior.
//
// Parameters:
//   - cfg: The application configuration containing the algorithm selection.
//   - factory: The calculator factory to retrieve implementations from.
//
// Returns:
//   - []fibonacci.Calculator: A slice of calculators to execute.
func GetCalculatorsToRun(cfg config.AppConfig, factory fibonacci.CalculatorFactory) []fibonacci.Calculator {
	if cfg.Algo == "all" {
		keys := factory.List() // List() returns sorted keys
		calculators := make([]fibonacci.Calculator, 0, len(keys))
		for _, k := range keys {
			if calc, err := factory.Get(k); err == nil {
				calculators = append(calculators, calc)
			}
		}
		return calculators
	}
	if calc, err := factory.Get(cfg.Algo); err == nil {
		return []fibonacci.Calculator{calc}
	}
	return nil
}
