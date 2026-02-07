// Number formatting utilities for CLI output.

package cli

import "strings"

// FormatNumberString inserts thousand separators into a numeric string.
// Optimized to reduce memory allocations
//
// Parameters:
//   - s: The numeric string to format.
//
// Returns:
//   - string: The formatted string with comma separators.
func FormatNumberString(s string) string {
	if s == "" {
		return ""
	}
	prefix := ""
	if s[0] == '-' {
		prefix = "-"
		s = s[1:]
	}
	n := len(s)
	if n <= 3 {
		return prefix + s
	}

	// Precise calculation of the required capacity to avoid reallocations
	numSeparators := (n - 1) / 3
	capacity := len(prefix) + n + numSeparators
	var builder strings.Builder
	builder.Grow(capacity)
	builder.WriteString(prefix)

	firstGroupLen := n % 3
	if firstGroupLen == 0 {
		firstGroupLen = 3
	}
	builder.WriteString(s[:firstGroupLen])

	// Optimized loop with fewer function calls
	for i := firstGroupLen; i < n; i += 3 {
		builder.WriteByte(',')
		builder.WriteString(s[i : i+3])
	}
	return builder.String()
}
