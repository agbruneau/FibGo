package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateAlgorithms handles key messages for the algorithms section.
func (m DashboardModel) updateAlgorithms(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.algorithms.cursor > 0 {
			m.algorithms.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.algorithms.cursor < len(m.algorithms.names)-1 {
			m.algorithms.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		// Start calculation with selected algorithm
		return m.startSingleCalculation()
	}
	return m, nil
}

// Column widths for the algorithm table (shared between header and rows).
const (
	colWidthRank     = 3
	colWidthName     = 25
	colWidthProgress = 30
	colWidthPct      = 8
	colWidthDur      = 12
	colWidthStatus   = 8
)

// calculateTableWidth returns the total width of the algorithm table row.
func calculateTableWidth() int {
	// 2 (left pad) + rank + 1 + name + 1 + progress + 1 + pct + 1 + dur + 1 + status
	return 2 + colWidthRank + 1 + colWidthName + 1 + colWidthProgress + 1 + colWidthPct + 1 + colWidthDur + 1 + colWidthStatus
}

// renderAlgorithmTable renders the algorithm comparison table.
func (m DashboardModel) renderAlgorithmTable() string {
	var b strings.Builder

	// Section title
	titleStyle := m.styles.BoxTitle
	if m.focusedSection == SectionAlgorithms {
		titleStyle = titleStyle.Foreground(m.styles.Primary.GetForeground())
	}
	b.WriteString(titleStyle.Render("ALGORITHMS"))
	b.WriteString("\n\n")

	// Table header - using consistent column widths
	colRank := lipgloss.NewStyle().Width(colWidthRank)
	colName := lipgloss.NewStyle().Width(colWidthName)
	colProgress := lipgloss.NewStyle().Width(colWidthProgress)
	colPct := lipgloss.NewStyle().Width(colWidthPct).Align(lipgloss.Right)
	colDur := lipgloss.NewStyle().Width(colWidthDur).Align(lipgloss.Right)
	colStatus := lipgloss.NewStyle().Width(colWidthStatus).Align(lipgloss.Center)

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		"  ",
		colRank.Render("#"),
		" ",
		colName.Render("Algorithm"),
		" ",
		colProgress.Render("Progress"),
		" ",
		colPct.Render("%"),
		" ",
		colDur.Render("Duration"),
		" ",
		colStatus.Render("Status"),
	)
	b.WriteString(m.styles.TableHeader.Render(header))
	b.WriteString("\n")

	// Separator - match table width exactly
	separatorWidth := calculateTableWidth()
	b.WriteString(m.styles.Primary.Render(strings.Repeat("━", separatorWidth)))
	b.WriteString("\n")

	// Algorithm rows
	for i, name := range m.algorithms.names {
		row := m.renderAlgorithmRow(i, name)
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

// renderAlgorithmRow renders a single algorithm row.
func (m DashboardModel) renderAlgorithmRow(idx int, name string) string {
	progress := m.algorithms.progresses[idx]
	duration := m.algorithms.durations[idx]
	status := m.algorithms.statuses[idx]

	// Row style with alternating backgrounds
	rowStyle := m.styles.TableRow
	if idx%2 == 1 {
		rowStyle = m.styles.TableRowAlt
	}
	if m.focusedSection == SectionAlgorithms && idx == m.algorithms.cursor {
		rowStyle = m.styles.MenuItemActive
	}

	// Column styles with fixed widths (matching header)
	colRank := lipgloss.NewStyle().Width(colWidthRank)
	colName := lipgloss.NewStyle().Width(colWidthName)
	colPct := lipgloss.NewStyle().Width(colWidthPct).Align(lipgloss.Right)
	colDur := lipgloss.NewStyle().Width(colWidthDur).Align(lipgloss.Right)
	colStatus := lipgloss.NewStyle().Width(colWidthStatus).Align(lipgloss.Center)

	// Rank column - keep as plain text, apply color separately
	rank := fmt.Sprintf("%d", idx+1)
	isWinner := false
	if status == StatusComplete && m.results.hasResults && len(m.results.results) > 0 {
		// Find position in results (sorted by duration)
		for pos, r := range m.results.results {
			if r.Name == name {
				rank = fmt.Sprintf("%d", pos+1)
				if pos == 0 {
					isWinner = true
				}
				break
			}
		}
	}

	// Truncate algorithm name to fit column
	displayName := truncateString(name, colWidthName)

	// Progress bar (matching column width)
	bar := m.renderProgressBar(progress, colWidthProgress)

	// Percentage
	pct := fmt.Sprintf("%.1f%%", progress*100)

	// Duration column
	durStr := "-"
	if status == StatusComplete {
		durStr = formatDuration(duration)
	} else if status == StatusRunning && m.calculation.active {
		durStr = "..."
	}

	// Status indicator - render plain text first, then apply style
	// This ensures width calculation works correctly
	var statusText string
	var statusStyle lipgloss.Style
	switch status {
	case StatusIdle:
		statusText = "IDLE"
		statusStyle = colStatus.Foreground(m.styles.Muted.GetForeground())
	case StatusRunning:
		statusText = "RUN"
		statusStyle = colStatus.Foreground(m.styles.Info.GetForeground())
	case StatusComplete:
		statusText = "OK"
		statusStyle = colStatus.Foreground(m.styles.Success.GetForeground())
	case StatusError:
		statusText = "ERR"
		statusStyle = colStatus.Foreground(m.styles.Error.GetForeground())
	}

	// Render rank with proper styling
	rankStyle := colRank
	if isWinner {
		rankStyle = colRank.Foreground(m.styles.Success.GetForeground())
	}

	// Build row using lipgloss for proper alignment
	row := lipgloss.JoinHorizontal(lipgloss.Center,
		"  ",
		rankStyle.Render(rank),
		" ",
		colName.Render(displayName),
		" ",
		bar,
		" ",
		colPct.Render(pct),
		" ",
		colDur.Render(durStr),
		" ",
		statusStyle.Render(statusText),
	)

	return rowStyle.Render(row)
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// renderProgressBar renders a progress bar with exact width.
func (m DashboardModel) renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", width-filled)

	// Combine without additional styling to preserve exact width
	bar := m.styles.ProgressFilled.Render(filledStr) + m.styles.ProgressEmpty.Render(emptyStr)

	// Wrap in a fixed-width container to ensure alignment
	return lipgloss.NewStyle().Width(width).Render(bar)
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fµs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}
