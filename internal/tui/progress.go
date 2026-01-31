package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderProgress() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Progress"))
	b.WriteString("\n")

	if m.state == StateConfig {
		b.WriteString(dimStyle.Render("  Press Enter to start calculation..."))
		return panelStyle.Width(m.width - 2).Render(b.String())
	}

	barWidth := m.width - 38
	if barWidth < 10 {
		barWidth = 10
	}

	for i, algo := range m.algos {
		colorIdx := i % len(barColors)
		bar := renderBar(algo.Progress, barWidth, barColors[colorIdx])

		pct := fmt.Sprintf("%6.1f%%", algo.Progress*100)
		pctStyle := lipgloss.NewStyle().Foreground(barColors[colorIdx])

		name := lipgloss.NewStyle().
			Width(18).
			Foreground(barColors[colorIdx]).
			Render(algo.Name)

		if algo.Done && algo.Err != nil {
			b.WriteString(fmt.Sprintf("  %s [%s] %s\n",
				name, bar, errorStyle.Render("ERROR")))
		} else if algo.Done {
			b.WriteString(fmt.Sprintf("  %s [%s] %s\n",
				name, bar, pctStyle.Render(" done")))
		} else {
			b.WriteString(fmt.Sprintf("  %s [%s] %s\n",
				name, bar, pctStyle.Render(pct)))
		}
	}

	// ETA line
	if m.state == StateRunning {
		elapsed := time.Since(m.startTime)
		avgProgress := m.averageProgress()
		etaStr := formatETA(avgProgress, elapsed)
		b.WriteString(fmt.Sprintf("\n  %s %s",
			labelStyle.Render("Elapsed:"),
			accentCyanStyle.Render(formatDuration(elapsed))))
		b.WriteString(fmt.Sprintf("    %s %s",
			labelStyle.Render("ETA:"),
			accentYellowStyle.Render(etaStr)))
	}

	return panelStyle.Width(m.width - 2).Render(b.String())
}

func renderBar(progress float64, width int, c lipgloss.Color) string {
	if progress > 1.0 {
		progress = 1.0
	}
	if progress < 0.0 {
		progress = 0.0
	}
	filled := int(progress * float64(width))
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(c)
	emptyStyle := lipgloss.NewStyle().Foreground(colorBarBg)

	return filledStyle.Render(strings.Repeat("│", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))
}

func (m Model) averageProgress() float64 {
	if len(m.algos) == 0 {
		return 0
	}
	var total float64
	for _, a := range m.algos {
		total += a.Progress
	}
	return total / float64(len(m.algos))
}

func formatETA(progress float64, elapsed time.Duration) string {
	if progress <= 0.01 {
		return "calculating..."
	}
	if progress >= 1.0 {
		return "done"
	}
	remaining := time.Duration(float64(elapsed) * (1.0 - progress) / progress)
	return formatDuration(remaining)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1s"
	}
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%02dm", h, m)
}
