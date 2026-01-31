package tui

import "fmt"

func (m Model) renderStatusBar() string {
	var keys string

	switch m.state {
	case StateConfig:
		keys = fmt.Sprintf(
			" %s quit  %s start  %s next field  %s change algo",
			keyStyle.Render("q"),
			keyStyle.Render("enter"),
			keyStyle.Render("tab"),
			keyStyle.Render("←→"),
		)
	case StateRunning:
		keys = fmt.Sprintf(
			" %s cancel & quit",
			keyStyle.Render("q/ctrl+c"),
		)
	case StateResults:
		keys = fmt.Sprintf(
			" %s quit  %s new calculation",
			keyStyle.Render("q"),
			keyStyle.Render("r"),
		)
	}

	status := dimStyle.Render(fmt.Sprintf("  [%s]", m.statusMsg))

	return statusBarStyle.Render(keys + status)
}
