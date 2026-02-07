// Number formatting utilities for CLI output.

package cli

import "github.com/agbru/fibcalc/internal/format"

// FormatNumberString delegates to format.FormatNumberString.
func FormatNumberString(s string) string {
	return format.FormatNumberString(s)
}
