package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// updateResults handles key messages for the results section.
func (m DashboardModel) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.results.hasResults {
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.results.cursor > 0 {
			m.results.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.results.cursor < len(m.results.results)-1 {
			m.results.cursor++
		}
	}
	return m, nil
}

// renderResultsSection renders the results section of the dashboard.
func (m DashboardModel) renderResultsSection() string {
	var b strings.Builder

	// Section title
	titleStyle := m.styles.BoxTitle
	if m.focusedSection == SectionResults {
		titleStyle = titleStyle.Foreground(m.styles.Primary.GetForeground())
	}
	b.WriteString(titleStyle.Render("RESULT"))
	b.WriteString("\n\n")

	if !m.results.hasResults {
		b.WriteString(m.styles.Muted.Render("  No results yet. Press [c] to calculate or [m] to compare."))
		return b.String()
	}

	// Get the best result
	var bestResult string
	var bestAlgo string
	var bestDuration string
	var bitCount int

	if len(m.results.results) > 0 {
		for _, r := range m.results.results {
			if r.Err == nil && r.Result != nil {
				if bestResult == "" {
					bestResult = r.Result.String()
					bestAlgo = r.Name
					bestDuration = formatDuration(r.Duration)
					bitCount = r.Result.BitLen()
				}
				break
			}
		}
	}

	if bestResult == "" {
		b.WriteString(m.styles.Error.Render("  All calculations failed."))
		return b.String()
	}

	// Default view: show Global Status like CLI mode
	if !m.results.showDetails {
		// Global Status line
		if len(m.results.results) > 1 {
			if m.results.consistent {
				b.WriteString(fmt.Sprintf("  %s %s\n",
					m.styles.Success.Bold(true).Render("Global Status: Success."),
					m.styles.Success.Render("All valid results are consistent."),
				))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s\n",
					m.styles.Error.Bold(true).Render("Global Status: CRITICAL ERROR!"),
					m.styles.Error.Render("Inconsistency detected between results."),
				))
			}
		} else {
			b.WriteString(fmt.Sprintf("  %s\n",
				m.styles.Success.Bold(true).Render("Global Status: Success."),
			))
		}

		// Result binary size line (like CLI)
		b.WriteString(fmt.Sprintf("  Result binary size: %s bits.\n",
			m.styles.ResultValue.Render(formatNumber(bitCount)),
		))

		// Stats line
		b.WriteString(fmt.Sprintf("\n  Fastest: %s (%s)",
			m.styles.Success.Render(bestAlgo),
			m.styles.Muted.Render(bestDuration),
		))

		// Actions hint
		b.WriteString("\n\n")
		b.WriteString("  ")
		b.WriteString(m.styles.HelpKey.Render("[d]"))
		b.WriteString(m.styles.HelpDesc.Render(" Show details  "))
		b.WriteString(m.styles.HelpKey.Render("[Ctrl+S]"))
		b.WriteString(m.styles.HelpDesc.Render(" Save"))

		return b.String()
	}

	// Detailed view: show full result
	displayResult := formatResultValue(bestResult, m.results.showHex, m.results.showFull, m.getMaxValueLength())
	digitCount := len(bestResult)

	// Result line
	b.WriteString(fmt.Sprintf("  %s = %s  (%s digits, %s bits)\n",
		m.styles.Primary.Render(fmt.Sprintf("F(%d)", m.results.n)),
		m.styles.ResultValue.Render(displayResult),
		m.styles.Info.Render(formatNumber(digitCount)),
		m.styles.Info.Render(formatNumber(bitCount)),
	))

	// Stats line
	b.WriteString(fmt.Sprintf("  Fastest: %s (%s)",
		m.styles.Success.Render(bestAlgo),
		m.styles.Muted.Render(bestDuration),
	))

	// Consistency check
	if len(m.results.results) > 1 {
		if m.results.consistent {
			b.WriteString(m.styles.Success.Bold(true).Render("   [OK] All results consistent"))
		} else {
			b.WriteString(m.styles.Error.Bold(true).Render("   [!] Results inconsistent!"))
		}
	}

	// Actions hint
	b.WriteString("\n\n")
	b.WriteString("  ")
	b.WriteString(m.styles.HelpKey.Render("[d]"))
	b.WriteString(m.styles.HelpDesc.Render(" Hide details  "))
	b.WriteString(m.styles.HelpKey.Render("[x]"))
	b.WriteString(m.styles.HelpDesc.Render(" Toggle hex  "))
	b.WriteString(m.styles.HelpKey.Render("[v]"))
	b.WriteString(m.styles.HelpDesc.Render(" Toggle full  "))
	b.WriteString(m.styles.HelpKey.Render("[Ctrl+S]"))
	b.WriteString(m.styles.HelpDesc.Render(" Save"))

	return b.String()
}

// formatResultValue formats the result value for display.
func formatResultValue(result string, showHex, showFull bool, maxLen int) string {
	if showFull {
		return result
	}

	if len(result) <= maxLen {
		return result
	}

	// Truncate with ellipsis
	half := (maxLen - 3) / 2
	return result[:half] + "..." + result[len(result)-half:]
}

// formatNumber formats a number with thousand separators.
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	remainder := len(str) % 3
	if remainder > 0 {
		result.WriteString(str[:remainder])
		if len(str) > remainder {
			result.WriteString(",")
		}
	}

	for i := remainder; i < len(str); i += 3 {
		result.WriteString(str[i : i+3])
		if i+3 < len(str) {
			result.WriteString(",")
		}
	}

	return result.String()
}

// getMaxValueLength returns the maximum result value length based on terminal width.
func (m DashboardModel) getMaxValueLength() int {
	if m.width > 160 {
		return 120
	}
	if m.width > 120 {
		return 80
	}
	if m.width > 80 {
		return 50
	}
	return 30
}
