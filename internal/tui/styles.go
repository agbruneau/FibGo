package tui

import "github.com/charmbracelet/lipgloss"

// HTOP-inspired color palette.
var (
	colorCyan    = lipgloss.Color("#00d2ff")
	colorGreen   = lipgloss.Color("#00e676")
	colorMagenta = lipgloss.Color("#e040fb")
	colorYellow  = lipgloss.Color("#ffea00")
	colorRed     = lipgloss.Color("#ff5252")
	colorDim     = lipgloss.Color("#555555")
	colorBarBg   = lipgloss.Color("#333333")
	colorBorder  = lipgloss.Color("#30475e")
	colorWhite   = lipgloss.Color("#e0e0e0")
	colorLabel   = lipgloss.Color("#90a4ae")

	// Bar colors per algorithm slot.
	barColors = []lipgloss.Color{colorGreen, colorCyan, colorMagenta}
)

// Panel styles.
var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorLabel)

	accentCyanStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	accentGreenStyle = lipgloss.NewStyle().
				Foreground(colorGreen)

	accentYellowStyle = lipgloss.NewStyle().
				Foreground(colorYellow)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	focusedStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	keyStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	resultRankStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed)
)
