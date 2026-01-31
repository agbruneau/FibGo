package tui

import (
	"testing"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/ui"
)

func TestNewModel(t *testing.T) {
	cfg := config.AppConfig{
		N:       1000,
		Algo:    "all",
		Timeout: config.DefaultTimeout,
	}

	// Create a minimal model without calculators for testing
	model := NewModel(cfg, nil)

	if model.currentView != ViewHome {
		t.Errorf("expected initial view to be ViewHome, got %v", model.currentView)
	}

	if model.config.N != 1000 {
		t.Errorf("expected N to be 1000, got %d", model.config.N)
	}
}

func TestDefaultStyles(t *testing.T) {
	// Test dark theme
	ui.SetTheme("dark")
	styles := DefaultStyles()
	// Verify styles are created without panicking
	_ = styles.Primary.Render("test")
	_ = styles.Success.Render("test")

	// Test light theme
	ui.SetTheme("light")
	styles = DefaultStyles()
	_ = styles.Primary.Render("test")

	// Test no-color theme
	ui.SetTheme("none")
	styles = DefaultStyles()
	// No-color theme should still work
	rendered := styles.Bold.Render("test")
	if rendered == "" {
		t.Error("expected Bold.Render to return non-empty string")
	}
}

func TestKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that key bindings are defined
	if len(km.Quit.Keys()) == 0 {
		t.Error("expected Quit key binding to have keys defined")
	}

	if len(km.Enter.Keys()) == 0 {
		t.Error("expected Enter key binding to have keys defined")
	}

	// Test ShortHelp returns bindings
	shortHelp := km.ShortHelp()
	if len(shortHelp) == 0 {
		t.Error("expected ShortHelp to return bindings")
	}

	// Test FullHelp returns bindings
	fullHelp := km.FullHelp()
	if len(fullHelp) == 0 {
		t.Error("expected FullHelp to return bindings")
	}
}

func TestGetThemeColors(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantValid bool
	}{
		{"dark theme", "dark", true},
		{"light theme", "light", true},
		{"no-color theme", "none", true},
		{"unknown theme defaults to dark", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colors := getThemeColors(tt.themeName)
			// Just verify it doesn't panic and returns something
			_ = colors.primary
			_ = colors.secondary
		})
	}
}

func TestProgressBarRender(t *testing.T) {
	styles := DefaultStyles()

	tests := []struct {
		name     string
		progress float64
		width    int
	}{
		{"zero progress", 0.0, 20},
		{"half progress", 0.5, 20},
		{"full progress", 1.0, 20},
		{"over 100%", 1.5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := renderProgressBar(tt.progress, tt.width, styles)
			if bar == "" {
				t.Error("expected non-empty progress bar")
			}
			// Check that bar contains the progress characters
			if len(bar) == 0 {
				t.Error("expected progress bar to have content")
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{0, 0, 0},
		{-1, 1, 1},
		{-5, -3, -3},
	}

	for _, tt := range tests {
		got := maxInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("maxInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
