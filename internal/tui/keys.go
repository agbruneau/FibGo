package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the TUI.
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding

	// Actions
	Enter  key.Binding
	Escape key.Binding
	Help   key.Binding
	Quit   key.Binding

	// Shortcuts
	Calculate key.Binding
	Compare   key.Binding
	Theme     key.Binding
	Settings  key.Binding
	Save      key.Binding
	Hex       key.Binding
	Full      key.Binding
	Details   key.Binding

	// Legacy aliases (for backward compatibility)
	NewCalc    key.Binding
	SaveResult key.Binding
	HexToggle  key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	calcBinding := key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "calculate"),
	)
	saveBinding := key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	)
	hexBinding := key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "hex"),
	)

	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "move right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next section"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev section"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "f1"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Calculate: calcBinding,
		Compare: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "compare"),
		),
		Theme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "theme"),
		),
		Settings: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "settings"),
		),
		Save: saveBinding,
		Hex:  hexBinding,
		Full: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "full value"),
		),
		Details: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "details"),
		),
		// Legacy aliases
		NewCalc:    calcBinding,
		SaveResult: saveBinding,
		HexToggle:  hexBinding,
	}
}

// ShortHelp returns keybindings to show in the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.Escape, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.ShiftTab, k.Enter, k.Escape},
		{k.NewCalc, k.Compare, k.Theme, k.Settings},
		{k.SaveResult, k.HexToggle, k.Help, k.Quit},
	}
}
