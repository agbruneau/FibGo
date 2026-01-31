package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/agbru/fibcalc/internal/ui"
)

// Styles holds all the lipgloss styles for the TUI.
type Styles struct {
	// App frame
	App       lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style
	Content   lipgloss.Style
	StatusBar lipgloss.Style

	// Colors derived from theme
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	Info      lipgloss.Style

	// Components
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	MenuItem       lipgloss.Style
	MenuItemActive lipgloss.Style
	Input          lipgloss.Style
	InputFocused   lipgloss.Style
	Button         lipgloss.Style
	ButtonFocused  lipgloss.Style
	ProgressBar    lipgloss.Style
	ProgressFilled lipgloss.Style
	ProgressEmpty  lipgloss.Style
	Table          lipgloss.Style
	TableHeader    lipgloss.Style
	TableRow       lipgloss.Style
	TableRowAlt    lipgloss.Style
	HelpKey        lipgloss.Style
	HelpDesc       lipgloss.Style
	ResultValue    lipgloss.Style
	ResultLabel    lipgloss.Style
	Box            lipgloss.Style
	BoxTitle       lipgloss.Style
	Highlight      lipgloss.Style
	Muted          lipgloss.Style
	Bold           lipgloss.Style
}

// themeColors holds the color palette for a theme.
type themeColors struct {
	primary      lipgloss.TerminalColor
	secondary    lipgloss.TerminalColor
	success      lipgloss.TerminalColor
	warning      lipgloss.TerminalColor
	err          lipgloss.TerminalColor
	info         lipgloss.TerminalColor
	accent       lipgloss.TerminalColor
	bgSubtle     lipgloss.TerminalColor
	bgHighlight  lipgloss.TerminalColor
	bgTableAlt   lipgloss.TerminalColor
	bgInputFocus lipgloss.TerminalColor
}

func getThemeColors(themeName string) themeColors {
	switch themeName {
	case "light":
		return themeColors{
			primary:      lipgloss.Color("#0066DD"),
			secondary:    lipgloss.Color("#555555"),
			success:      lipgloss.Color("#228B22"),
			warning:      lipgloss.Color("#CC6600"),
			err:          lipgloss.Color("#CC0000"),
			info:         lipgloss.Color("#7722BB"),
			accent:       lipgloss.Color("#E83E8C"),
			bgSubtle:     lipgloss.Color("#F8F9FA"),
			bgHighlight:  lipgloss.Color("#E6F2FF"),
			bgTableAlt:   lipgloss.Color("#F0F4F8"),
			bgInputFocus: lipgloss.Color("#E6F0FF"),
		}
	case "none":
		return themeColors{
			primary:      lipgloss.NoColor{},
			secondary:    lipgloss.NoColor{},
			success:      lipgloss.NoColor{},
			warning:      lipgloss.NoColor{},
			err:          lipgloss.NoColor{},
			info:         lipgloss.NoColor{},
			accent:       lipgloss.NoColor{},
			bgSubtle:     lipgloss.NoColor{},
			bgHighlight:  lipgloss.NoColor{},
			bgTableAlt:   lipgloss.NoColor{},
			bgInputFocus: lipgloss.NoColor{},
		}
	default: // dark
		return themeColors{
			primary:      lipgloss.Color("#00D9FF"),
			secondary:    lipgloss.Color("#A0A0A0"),
			success:      lipgloss.Color("#00FF00"),
			warning:      lipgloss.Color("#FFCC00"),
			err:          lipgloss.Color("#FF5555"),
			info:         lipgloss.Color("#DD88FF"),
			accent:       lipgloss.Color("#FF6B9D"),
			bgSubtle:     lipgloss.Color("#1A1A2E"),
			bgHighlight:  lipgloss.Color("#0A2F4F"),
			bgTableAlt:   lipgloss.Color("#0D1B2A"),
			bgInputFocus: lipgloss.Color("#0A2540"),
		}
	}
}

// DefaultStyles creates styles based on the current theme.
func DefaultStyles() Styles {
	theme := ui.GetCurrentTheme()
	colors := getThemeColors(theme.Name)

	s := Styles{}
	s.initFrameStyles(colors)
	s.initColorStyles(colors)
	s.initComponentStyles(colors)
	return s
}

func (s *Styles) initFrameStyles(colors themeColors) {
	s.App = lipgloss.NewStyle().Padding(1, 2)
	s.Header = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		Padding(0, 1).
		MarginBottom(1)
	s.Footer = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(true).
		Padding(0, 1).
		MarginTop(1)
	s.Content = lipgloss.NewStyle().Padding(1, 2)
	s.StatusBar = lipgloss.NewStyle().Foreground(colors.secondary).Padding(0, 1)
}

func (s *Styles) initColorStyles(colors themeColors) {
	s.Primary = lipgloss.NewStyle().Foreground(colors.primary)
	s.Secondary = lipgloss.NewStyle().Foreground(colors.secondary)
	s.Success = lipgloss.NewStyle().Foreground(colors.success)
	s.Warning = lipgloss.NewStyle().Foreground(colors.warning)
	s.Error = lipgloss.NewStyle().Foreground(colors.err)
	s.Info = lipgloss.NewStyle().Foreground(colors.info)
}

func (s *Styles) initComponentStyles(colors themeColors) {
	s.Title = lipgloss.NewStyle().Foreground(colors.primary).Bold(true).MarginBottom(1)
	s.Subtitle = lipgloss.NewStyle().Foreground(colors.secondary).Italic(true)
	s.MenuItem = lipgloss.NewStyle().Padding(0, 2)
	s.MenuItemActive = lipgloss.NewStyle().
		Foreground(colors.primary).
		Background(colors.bgHighlight).
		Bold(true).
		Padding(0, 2)

	s.Input = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
	s.InputFocused = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colors.primary).
		Background(colors.bgInputFocus).
		Padding(0, 1)

	s.Button = lipgloss.NewStyle().
		Foreground(colors.secondary).
		Padding(0, 2).
		MarginRight(1)
	s.ButtonFocused = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colors.primary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)

	s.ProgressBar = lipgloss.NewStyle().Width(40)
	s.ProgressFilled = lipgloss.NewStyle().Foreground(colors.success)
	s.ProgressEmpty = lipgloss.NewStyle().Foreground(colors.secondary)

	s.Table = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
	s.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.primary).
		Background(colors.bgHighlight).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1)
	s.TableRow = lipgloss.NewStyle().Padding(0, 1)
	s.TableRowAlt = lipgloss.NewStyle().
		Background(colors.bgTableAlt).
		Padding(0, 1)

	s.HelpKey = lipgloss.NewStyle().Foreground(colors.primary).Bold(true)
	s.HelpDesc = lipgloss.NewStyle().Foreground(colors.secondary)

	s.ResultValue = lipgloss.NewStyle().Foreground(colors.success).Bold(true)
	s.ResultLabel = lipgloss.NewStyle().Foreground(colors.secondary)

	s.Box = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2)
	s.BoxTitle = lipgloss.NewStyle().Bold(true).Foreground(colors.primary)

	s.Highlight = lipgloss.NewStyle().Foreground(colors.accent).Bold(true)
	s.Muted = lipgloss.NewStyle().Foreground(colors.secondary)
	s.Bold = lipgloss.NewStyle().Bold(true)
}

// RefreshStyles updates styles when the theme changes.
func (s *Styles) RefreshStyles() {
	*s = DefaultStyles()
}
