package tui

import (
	"fmt"
	"strings"
)

func (m Model) renderConfig() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Configuration"))
	b.WriteString("\n")

	// N field
	nLabel := labelStyle.Render("  N:")
	if m.state == StateConfig && m.focus == FocusN {
		nLabel = focusedStyle.Render("▸ N:")
	}
	b.WriteString(fmt.Sprintf("%s         %s\n", nLabel, m.nInput.View()))

	// Algorithm selector
	algoLabel := labelStyle.Render("  Algorithm:")
	if m.state == StateConfig && m.focus == FocusAlgo {
		algoLabel = focusedStyle.Render("▸ Algorithm:")
	}
	var algoParts []string
	for i, a := range m.algoChoices {
		if i == m.algoIndex {
			algoParts = append(algoParts, accentGreenStyle.Render(fmt.Sprintf("◀ %s ▶", a)))
		} else {
			algoParts = append(algoParts, dimStyle.Render(a))
		}
	}
	b.WriteString(fmt.Sprintf("%s %s\n", algoLabel, strings.Join(algoParts, "  ")))

	// Timeout field
	tLabel := labelStyle.Render("  Timeout:")
	if m.state == StateConfig && m.focus == FocusTimeout {
		tLabel = focusedStyle.Render("▸ Timeout:")
	}
	b.WriteString(fmt.Sprintf("%s   %s", tLabel, m.timeoutInput.View()))

	return panelStyle.Width(m.width - 2).Render(b.String())
}
