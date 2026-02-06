package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/agbru/fibcalc/internal/cli"
)

// ChartModel renders a progress bar, ETA, and system metrics sparklines.
type ChartModel struct {
	averageProgress float64
	eta             time.Duration
	elapsed         time.Duration
	done            bool
	width           int
	height          int

	cpuHistory *RingBuffer
	memHistory *RingBuffer
}

const defaultSparklineCap = 30

// NewChartModel creates a new chart.
func NewChartModel() ChartModel {
	return ChartModel{
		cpuHistory: NewRingBuffer(defaultSparklineCap),
		memHistory: NewRingBuffer(defaultSparklineCap),
	}
}

// SetSize updates dimensions.
func (c *ChartModel) SetSize(w, h int) {
	c.width = w
	c.height = h
	if sw := c.sparklineWidth(); sw > 0 {
		c.cpuHistory.Resize(sw)
		c.memHistory.Resize(sw)
	}
}

// AddDataPoint records a progress sample.
func (c *ChartModel) AddDataPoint(progress float64, avg float64, eta time.Duration) {
	c.averageProgress = avg
	c.eta = eta
}

// UpdateSysStats records a system metrics sample.
func (c *ChartModel) UpdateSysStats(cpuPct, memPct float64) {
	c.cpuHistory.Push(cpuPct)
	c.memHistory.Push(memPct)
}

// SetDone marks the chart as complete with the total elapsed time.
func (c *ChartModel) SetDone(elapsed time.Duration) {
	c.done = true
	c.averageProgress = 1.0
	c.elapsed = elapsed
}

// Reset clears the chart data.
func (c *ChartModel) Reset() {
	c.averageProgress = 0
	c.eta = 0
	c.elapsed = 0
	c.done = false
	c.cpuHistory.Reset()
	c.memHistory.Reset()
}

// View renders the chart panel.
func (c ChartModel) View() string {
	var b strings.Builder

	b.WriteString(metricLabelStyle.Render("  Progress Chart"))
	b.WriteString("\n\n")

	// Render progress bar
	progressBar := c.renderProgressBar()
	if progressBar != "" {
		b.WriteString("  ")
		b.WriteString(progressBar)
		b.WriteString("\n\n")
	}

	// Render ETA or total elapsed time
	var statusStr string
	if c.done {
		statusStr = fmt.Sprintf("Completed in %s", cli.FormatExecutionDuration(c.elapsed))
	} else {
		statusStr = fmt.Sprintf("ETA: %s", cli.FormatETA(c.eta))
	}
	b.WriteString(fmt.Sprintf("  %s", elapsedStyle.Render(statusStr)))

	// Render system sparklines if space allows
	if c.height >= 10 && c.sparklineWidth() > 0 {
		b.WriteString("\n\n")
		b.WriteString(c.renderSparklineLine("CPU", c.cpuHistory, cpuSparklineStyle))
		b.WriteString("\n")
		b.WriteString(c.renderSparklineLine("MEM", c.memHistory, memSparklineStyle))
	}

	return panelStyle.
		Width(c.width - 2).
		Height(c.height - 2).
		Render(b.String())
}

func (c ChartModel) renderProgressBar() string {
	barWidth := c.width - 15 // border + indent + brackets + " 100.0%"
	if barWidth < 5 {
		return ""
	}

	filled := int(c.averageProgress * float64(barWidth))
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	filledStr := chartBarStyle.Render(strings.Repeat("█", filled))
	emptyStr := chartEmptyStyle.Render(strings.Repeat("░", empty))
	pctStr := metricValueStyle.Render(fmt.Sprintf("%5.1f%%", c.averageProgress*100))

	return fmt.Sprintf("[%s%s] %s", filledStr, emptyStr, pctStr)
}

// sparklineWidth computes the number of characters available for the sparkline.
// Line format: "  CPU  xxx.x%  <sparkline>"  (~15 chars for label+pct, +2 border)
func (c ChartModel) sparklineWidth() int {
	w := c.width - 17
	if w < 1 {
		return 0
	}
	return w
}

func (c ChartModel) renderSparklineLine(label string, buf *RingBuffer, style lipgloss.Style) string {
	pct := buf.Last()
	pctStr := metricValueStyle.Render(fmt.Sprintf("%5.1f%%", pct))
	sparkStr := style.Render(RenderSparkline(buf.Slice()))
	return fmt.Sprintf("  %s  %s  %s",
		metricLabelStyle.Render(label),
		pctStr,
		sparkStr)
}
