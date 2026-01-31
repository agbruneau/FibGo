package cli

import "regexp"

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripAnsiCodes removes ANSI escape codes from a string.
func stripAnsiCodes(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}
