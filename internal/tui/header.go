package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/agbru/fibcalc/internal/cli"
)

// HeaderModel renders the top bar: title, version, elapsed time.
type HeaderModel struct {
	startTime time.Time
	endTime   time.Time
	version   string
	width     int
}

// NewHeaderModel creates a new header.
func NewHeaderModel(version string) HeaderModel {
	return HeaderModel{
		startTime: time.Now(),
		version:   version,
	}
}

// SetDone freezes the elapsed timer at the current time.
func (h *HeaderModel) SetDone() {
	h.endTime = time.Now()
}

// SetWidth updates the available width.
func (h *HeaderModel) SetWidth(w int) {
	h.width = w
}

// View renders the header.
func (h HeaderModel) View() string {
	title := titleStyle.Render("FibGo Monitor")
	version := versionStyle.Render(h.version)
	var duration time.Duration
	if !h.endTime.IsZero() {
		duration = h.endTime.Sub(h.startTime)
	} else {
		duration = time.Since(h.startTime)
	}
	elapsed := elapsedStyle.Render(fmt.Sprintf("Elapsed: %s", cli.FormatExecutionDuration(duration)))

	// Calculate spacing
	titleLen := lipgloss.Width(title)
	versionLen := lipgloss.Width(version)
	elapsedLen := lipgloss.Width(elapsed)

	innerWidth := h.width - 4 // account for panel borders/padding
	if innerWidth < 0 {
		innerWidth = 0
	}

	gap := innerWidth - titleLen - versionLen - elapsedLen
	if gap < 2 {
		gap = 2
	}
	leftGap := gap / 2
	rightGap := gap - leftGap

	row := title + spaces(leftGap) + version + spaces(rightGap) + elapsed

	return headerStyle.Width(innerWidth).Render(row)
}

// spaces returns a string of n space characters.
func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
