// Package ui provides terminal color utilities with NO_COLOR support.
package ui

import "os"

// noColor indicates whether color output is disabled.
var noColor bool

func init() {
	// Respect NO_COLOR environment variable (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		noColor = true
	}
}

// InitTheme initializes color support. When noColorFlag is true, all color
// functions return empty strings.
func InitTheme(noColorFlag bool) {
	noColor = noColorFlag
}

func color(code string) string {
	if noColor {
		return ""
	}
	return code
}

// ColorReset returns the ANSI reset code.
func ColorReset() string { return color("\033[0m") }

// ColorRed returns the ANSI red color code.
func ColorRed() string { return color("\033[31m") }

// ColorGreen returns the ANSI green color code.
func ColorGreen() string { return color("\033[32m") }

// ColorYellow returns the ANSI yellow color code.
func ColorYellow() string { return color("\033[33m") }

// ColorBlue returns the ANSI blue color code.
func ColorBlue() string { return color("\033[34m") }

// ColorMagenta returns the ANSI magenta color code.
func ColorMagenta() string { return color("\033[35m") }

// ColorCyan returns the ANSI cyan color code.
func ColorCyan() string { return color("\033[36m") }

// ColorBold returns the ANSI bold code.
func ColorBold() string { return color("\033[1m") }

// ColorUnderline returns the ANSI underline code.
func ColorUnderline() string { return color("\033[4m") }
