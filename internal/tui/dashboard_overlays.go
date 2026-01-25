package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/agbru/fibcalc/internal/ui"
)

// renderHelpOverlay renders the help overlay on top of the dashboard.
func (m DashboardModel) renderHelpOverlay() string {
	// Build help content
	helpContent := m.buildHelpContent()

	// Calculate overlay dimensions
	overlayWidth := min(70, m.width-4)
	overlayHeight := min(25, m.height-4)

	// Create overlay box style
	overlayStyle := lipgloss.NewStyle().
		Width(overlayWidth).
		Height(overlayHeight).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(m.styles.Primary.GetForeground()).
		Padding(1, 2).
		Align(lipgloss.Left)

	// Render overlay
	overlay := overlayStyle.Render(helpContent)

	// Center the overlay on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}

// buildHelpContent builds the help overlay content.
func (m DashboardModel) buildHelpContent() string {
	var b strings.Builder

	// Title
	b.WriteString(m.styles.Title.Render("FIBONACCI CALCULATOR - HELP"))
	b.WriteString("\n\n")

	// Navigation section
	b.WriteString(m.styles.BoxTitle.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(formatHelpLine(m.styles, "Tab / Shift+Tab", "Cycle through sections"))
	b.WriteString(formatHelpLine(m.styles, "Up/Down / k/j", "Navigate within section"))
	b.WriteString(formatHelpLine(m.styles, "Enter", "Confirm / Start calculation"))
	b.WriteString(formatHelpLine(m.styles, "Esc", "Cancel / Close overlay"))
	b.WriteString("\n")

	// Actions section
	b.WriteString(m.styles.BoxTitle.Render("Actions"))
	b.WriteString("\n")
	b.WriteString(formatHelpLine(m.styles, "c", "Calculate F(N) with selected algorithm"))
	b.WriteString(formatHelpLine(m.styles, "m", "Compare all algorithms"))
	b.WriteString(formatHelpLine(m.styles, "x", "Toggle hexadecimal display"))
	b.WriteString(formatHelpLine(m.styles, "v", "Toggle full value display"))
	b.WriteString(formatHelpLine(m.styles, "Ctrl+S", "Save result to file"))
	b.WriteString("\n")

	// UI section
	b.WriteString(m.styles.BoxTitle.Render("Interface"))
	b.WriteString("\n")
	b.WriteString(formatHelpLine(m.styles, "t", "Cycle theme (dark/light/none)"))
	b.WriteString(formatHelpLine(m.styles, "? / F1", "Toggle this help"))
	b.WriteString(formatHelpLine(m.styles, "q / Ctrl+C", "Quit"))
	b.WriteString("\n")

	// About section
	b.WriteString(m.styles.Muted.Render(strings.Repeat("-", 50)))
	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render("FibCalc - High-Performance Fibonacci Calculator"))
	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render("Press ? or Esc to close this help"))

	return b.String()
}

// formatHelpLine formats a help line with key and description.
func formatHelpLine(styles Styles, key, desc string) string {
	return fmt.Sprintf("  %s  %s\n",
		styles.HelpKey.Width(15).Render(key),
		styles.HelpDesc.Render(desc),
	)
}

// renderDashboard renders the main dashboard view.
func (m DashboardModel) renderDashboard() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Error bar (if any)
	if m.lastError != nil {
		errorBar := m.styles.Error.Render(fmt.Sprintf("  Error: %v", m.lastError))
		sections = append(sections, errorBar)
	}

	// Input section
	inputSection := m.renderInputSection()
	inputBox := m.styles.Box.Width(m.width - 4).Render(inputSection)
	sections = append(sections, inputBox)

	// Algorithms section
	algoSection := m.renderAlgorithmTable()
	algoBox := m.styles.Box.Width(m.width - 4).Render(algoSection)
	sections = append(sections, algoBox)

	// Results section
	resultsSection := m.renderResultsSection()
	resultsBox := m.styles.Box.Width(m.width - 4).Render(resultsSection)
	sections = append(sections, resultsBox)

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the dashboard header.
func (m DashboardModel) renderHeader() string {
	theme := ui.GetCurrentTheme()

	left := m.styles.Title.Render("FIBONACCI CALCULATOR")
	rightFull := fmt.Sprintf("Theme: %s  [?] Help", theme.Name)
	rightShort := "[?] Help"

	// Calculate available width (accounting for header padding)
	availableWidth := m.width - 6
	if availableWidth < 20 {
		availableWidth = 20
	}

	leftWidth := lipgloss.Width(left)
	rightFullWidth := lipgloss.Width(m.styles.Muted.Render(rightFull))
	rightShortWidth := lipgloss.Width(m.styles.Muted.Render(rightShort))

	// Choose right text based on available space
	var right string
	var spacing int

	if leftWidth+rightFullWidth+2 <= availableWidth {
		// Full version fits
		right = m.styles.Muted.Render(rightFull)
		spacing = availableWidth - leftWidth - rightFullWidth
	} else if leftWidth+rightShortWidth+2 <= availableWidth {
		// Short version fits
		right = m.styles.Muted.Render(rightShort)
		spacing = availableWidth - leftWidth - rightShortWidth
	} else {
		// Nothing fits on right, just show title
		right = ""
		spacing = 0
	}

	// Ensure spacing is never negative
	if spacing < 0 {
		spacing = 0
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		left,
		strings.Repeat(" ", spacing),
		right,
	)

	return m.styles.Header.Width(m.width - 2).Render(header)
}

// renderFooter renders the dashboard footer with context-sensitive help.
func (m DashboardModel) renderFooter() string {
	var hints []string

	// Context-sensitive hints based on focused section
	switch m.focusedSection {
	case SectionInput:
		hints = append(hints, "Enter:Submit", "Tab:Next")
	case SectionAlgorithms:
		hints = append(hints, "Enter:Run", "Up/Down:Select", "Tab:Next")
	case SectionResults:
		if m.results.showDetails {
			hints = append(hints, "d:Hide", "x:Hex", "v:Full", "Ctrl+S:Save", "Tab:Next")
		} else {
			hints = append(hints, "d:Details", "Ctrl+S:Save", "Tab:Next")
		}
	}

	// Global hints
	hints = append(hints, "c:Calc", "m:Compare", "t:Theme", "?:Help", "q:Quit")

	// Join hints and truncate if too long
	footer := strings.Join(hints, "  ")

	// Ensure footer doesn't exceed available width
	maxWidth := m.width - 6
	if maxWidth < 20 {
		maxWidth = 20
	}
	if len(footer) > maxWidth {
		footer = footer[:maxWidth-3] + "..."
	}

	return m.styles.Footer.Width(m.width - 2).Render(footer)
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
