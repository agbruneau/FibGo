package tui

import "github.com/charmbracelet/lipgloss"

// Orange-dominant dark theme palette inspired by btop.
var (
	colorBg      = lipgloss.Color("#000000")
	colorText    = lipgloss.Color("#E0E0E0")
	colorBorder  = lipgloss.Color("#FF6600")
	colorAccent  = lipgloss.Color("#FF8C00")
	colorSuccess = lipgloss.Color("#9ece6a")
	colorWarning = lipgloss.Color("#FFB347")
	colorError   = lipgloss.Color("#FF4444")
	colorDim     = lipgloss.Color("#666666")
	colorCyan    = lipgloss.Color("#FF8C00")
	colorMagenta = lipgloss.Color("#4488FF")
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

// chartBarStyle for the sparkline characters and filled progress bar.
var chartBarStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// chartEmptyStyle for the empty portion of the progress bar.
var chartEmptyStyle = lipgloss.NewStyle().
	Foreground(colorDim)

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

// cpuSparklineStyle for CPU sparkline characters (orange).
var cpuSparklineStyle = lipgloss.NewStyle().
	Foreground(colorAccent)

// memSparklineStyle for memory sparkline characters (warm orange).
var memSparklineStyle = lipgloss.NewStyle().
	Foreground(colorWarning)

