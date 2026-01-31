package tui

import (
	"fmt"
	"runtime"
)

func (m Model) renderHeader() string {
	version := m.version
	if version == "" {
		version = "dev"
	}

	content := fmt.Sprintf(
		"%s  │  %s  │  %s  │  %s",
		headerStyle.Render(fmt.Sprintf("FibGo %s", version)),
		accentCyanStyle.Render(fmt.Sprintf("%d CPUs", runtime.NumCPU())),
		accentCyanStyle.Render(runtime.Version()),
		dimStyle.Render(runtime.GOOS),
	)

	return panelStyle.Width(m.width - 2).Render(content)
}
