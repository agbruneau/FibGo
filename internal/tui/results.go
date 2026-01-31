package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/agbru/fibcalc/internal/cli"
)

func (m Model) renderResults() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Results"))
	b.WriteString("\n")

	if m.state == StateConfig {
		b.WriteString(dimStyle.Render("  No results yet."))
		return panelStyle.Width(m.width - 2).Render(b.String())
	}

	if m.state == StateRunning {
		completed := 0
		for _, a := range m.algos {
			if a.Done {
				completed++
			}
		}
		b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d completed...", completed, len(m.algos))))
		return panelStyle.Width(m.width - 2).Render(b.String())
	}

	// StateResults: show sorted results
	sorted := make([]AlgoProgress, len(m.algos))
	copy(sorted, m.algos)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Err != nil {
			return false
		}
		if sorted[j].Err != nil {
			return true
		}
		return sorted[i].Duration < sorted[j].Duration
	})

	for i, a := range sorted {
		rank := resultRankStyle.Render(fmt.Sprintf("#%d", i+1))
		colorIdx := i % len(barColors)
		nameStyle := lipgloss.NewStyle().Foreground(barColors[colorIdx]).Width(20)

		if a.Err != nil {
			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				rank,
				nameStyle.Render(a.Name),
				errorStyle.Render(fmt.Sprintf("error: %v", a.Err))))
			continue
		}

		dur := cli.FormatExecutionDuration(a.Duration)
		digits := formatNumberStr(fmt.Sprintf("%d", a.Digits))
		bits := formatNumberStr(fmt.Sprintf("%d", a.BitLen))

		b.WriteString(fmt.Sprintf("  %s %s %s   %s   %s\n",
			rank,
			nameStyle.Render(a.Name),
			accentGreenStyle.Render(fmt.Sprintf("%-10s", dur)),
			accentCyanStyle.Render(fmt.Sprintf("%s digits", digits)),
			dimStyle.Render(fmt.Sprintf("(%s bits)", bits))))
	}

	return panelStyle.Width(m.width - 2).Render(b.String())
}

// formatNumberStr inserts thousand separators.
func formatNumberStr(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var sb strings.Builder
	sb.Grow(n + (n-1)/3)
	first := n % 3
	if first == 0 {
		first = 3
	}
	sb.WriteString(s[:first])
	for i := first; i < n; i += 3 {
		sb.WriteByte(',')
		sb.WriteString(s[i : i+3])
	}
	return sb.String()
}
