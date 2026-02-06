package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/agbru/fibcalc/internal/cli"
)

// sparkBlocks are the braille/block characters used for the sparkline.
var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// ChartModel renders a sparkline chart of progress over time.
type ChartModel struct {
	dataPoints      []float64
	maxPoints       int
	averageProgress float64
	eta             time.Duration
	width           int
	height          int
}

// NewChartModel creates a new chart.
func NewChartModel() ChartModel {
	return ChartModel{
		dataPoints: make([]float64, 0, 64),
		maxPoints:  40,
	}
}

// SetSize updates dimensions and adjusts max data points.
func (c *ChartModel) SetSize(w, h int) {
	c.width = w
	c.height = h
	c.maxPoints = w - 6 // leave room for borders and padding
	if c.maxPoints < 5 {
		c.maxPoints = 5
	}
	// Trim excess data points
	if len(c.dataPoints) > c.maxPoints {
		c.dataPoints = c.dataPoints[len(c.dataPoints)-c.maxPoints:]
	}
}

// AddDataPoint records a progress sample.
func (c *ChartModel) AddDataPoint(progress float64, avg float64, eta time.Duration) {
	c.dataPoints = append(c.dataPoints, progress)
	if len(c.dataPoints) > c.maxPoints {
		c.dataPoints = c.dataPoints[len(c.dataPoints)-c.maxPoints:]
	}
	c.averageProgress = avg
	c.eta = eta
}

// Reset clears the chart data.
func (c *ChartModel) Reset() {
	c.dataPoints = c.dataPoints[:0]
	c.averageProgress = 0
	c.eta = 0
}

// View renders the chart panel.
func (c ChartModel) View() string {
	var b strings.Builder

	b.WriteString(metricLabelStyle.Render("  Progress Chart"))
	b.WriteString("\n\n")

	// Render sparkline
	sparkline := c.renderSparkline()
	b.WriteString("  ")
	b.WriteString(chartBarStyle.Render(sparkline))
	b.WriteString("\n\n")

	// Render progress bar
	progressBar := c.renderProgressBar()
	if progressBar != "" {
		b.WriteString("  ")
		b.WriteString(progressBar)
		b.WriteString("\n\n")
	}

	// Render stats
	avgStr := fmt.Sprintf("avg: %.1f%%", c.averageProgress*100)
	etaStr := fmt.Sprintf("ETA: %s", cli.FormatETA(c.eta))
	b.WriteString(fmt.Sprintf("  %s  %s",
		metricValueStyle.Render(avgStr),
		elapsedStyle.Render(etaStr)))

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

func (c ChartModel) renderSparkline() string {
	if len(c.dataPoints) == 0 {
		return ""
	}

	var b strings.Builder
	for _, v := range c.dataPoints {
		idx := int(v * float64(len(sparkBlocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkBlocks) {
			idx = len(sparkBlocks) - 1
		}
		b.WriteRune(sparkBlocks[idx])
	}
	return b.String()
}
