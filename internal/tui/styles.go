package tui

import "github.com/charmbracelet/lipgloss"

// btop-inspired dark theme palette.
var (
	colorBg      = lipgloss.Color("#1a1b26")
	colorText    = lipgloss.Color("#a9b1d6")
	colorBorder  = lipgloss.Color("#3b4261")
	colorAccent  = lipgloss.Color("#7aa2f7")
	colorSuccess = lipgloss.Color("#9ece6a")
	colorWarning = lipgloss.Color("#e0af68")
	colorError   = lipgloss.Color("#f7768e")
	colorDim     = lipgloss.Color("#565f89")
	colorCyan    = lipgloss.Color("#7dcfff")
	colorMagenta = lipgloss.Color("#bb9af7")
)

// panelStyle is the base style for bordered panels.
var panelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Background(colorBg).
	Foreground(colorText)

// headerStyle renders the top bar.
var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorAccent).
	Background(colorBg).
	Padding(0, 1)

// titleStyle for the FibGo Monitor title.
var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorAccent)

// versionStyle for the version label.
var versionStyle = lipgloss.NewStyle().
	Foreground(colorDim)

// elapsedStyle for the elapsed time.
var elapsedStyle = lipgloss.NewStyle().
	Foreground(colorCyan)

// logTimeStyle for timestamps in log entries.
var logTimeStyle = lipgloss.NewStyle().
	Foreground(colorDim)

// logAlgoStyle for algorithm names.
var logAlgoStyle = lipgloss.NewStyle().
	Foreground(colorMagenta)

// logProgressStyle for progress percentages.
var logProgressStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// logSuccessStyle for completed entries.
var logSuccessStyle = lipgloss.NewStyle().
	Foreground(colorSuccess)

// logErrorStyle for error entries.
var logErrorStyle = lipgloss.NewStyle().
	Foreground(colorError)

// metricLabelStyle for metric labels.
var metricLabelStyle = lipgloss.NewStyle().
	Foreground(colorDim)

// metricValueStyle for metric values.
var metricValueStyle = lipgloss.NewStyle().
	Foreground(colorCyan).
	Bold(true)

// chartBarStyle for the sparkline characters.
var chartBarStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// footerKeyStyle for keyboard shortcut keys.
var footerKeyStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

// footerDescStyle for keyboard shortcut descriptions.
var footerDescStyle = lipgloss.NewStyle().
	Foreground(colorDim)

// statusRunningStyle for Running indicator.
var statusRunningStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

// statusPausedStyle for Paused indicator.
var statusPausedStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

// statusDoneStyle for Done indicator.
var statusDoneStyle = lipgloss.NewStyle().
	Foreground(colorAccent).
	Bold(true)

// statusErrorStyle for Error indicator.
var statusErrorStyle = lipgloss.NewStyle().
	Foreground(colorError).
	Bold(true)

