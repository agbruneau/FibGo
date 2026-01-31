package tui

import (
	"context"
	"math/big"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/agbru/fibcalc/internal/config"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// mockCalculator implements the fibonacci.Calculator interface for testing.
type mockCalculator struct {
	name string
}

func (m *mockCalculator) Name() string { return m.name }
func (m *mockCalculator) Calculate(_ context.Context, _ chan<- fibonacci.ProgressUpdate, _ int, _ uint64, _ fibonacci.Options) (*big.Int, error) {
	return big.NewInt(0), nil
}

func newMockCalculators() []fibonacci.Calculator {
	return []fibonacci.Calculator{
		&mockCalculator{name: "fast-doubling"},
		&mockCalculator{name: "matrix"},
		&mockCalculator{name: "fft-hybrid"},
	}
}

func TestNewDashboardModel(t *testing.T) {
	cfg := config.AppConfig{N: 1000}
	calcs := newMockCalculators()

	model := NewDashboardModel(cfg, calcs)

	if model.config.N != 1000 {
		t.Errorf("expected N=1000, got %d", model.config.N)
	}
	if len(model.calculators) != 3 {
		t.Errorf("expected 3 calculators, got %d", len(model.calculators))
	}
	if model.focusedSection != SectionInput {
		t.Errorf("expected initial focus on SectionInput, got %v", model.focusedSection)
	}
	if model.input.n != "1000" {
		t.Errorf("expected input.n='1000', got '%s'", model.input.n)
	}
	if len(model.algorithms.names) != 3 {
		t.Errorf("expected 3 algorithm names, got %d", len(model.algorithms.names))
	}
}

func TestDashboardModel_SectionNavigation(t *testing.T) {
	cfg := config.AppConfig{N: 100}
	model := NewDashboardModel(cfg, newMockCalculators())
	model.ready = true
	model.width = 120
	model.height = 40

	// Initial state
	if model.focusedSection != SectionInput {
		t.Fatalf("expected SectionInput, got %v", model.focusedSection)
	}

	// Tab to next section
	msg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := model.Update(msg)
	model = result.(DashboardModel)
	if model.focusedSection != SectionAlgorithms {
		t.Errorf("after Tab, expected SectionAlgorithms, got %v", model.focusedSection)
	}

	// Tab again
	result, _ = model.Update(msg)
	model = result.(DashboardModel)
	if model.focusedSection != SectionResults {
		t.Errorf("after 2nd Tab, expected SectionResults, got %v", model.focusedSection)
	}

	// Tab wraps around
	result, _ = model.Update(msg)
	model = result.(DashboardModel)
	if model.focusedSection != SectionInput {
		t.Errorf("after 3rd Tab, expected SectionInput (wrap), got %v", model.focusedSection)
	}

	// Shift+Tab goes backward
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	result, _ = model.Update(msg)
	model = result.(DashboardModel)
	if model.focusedSection != SectionResults {
		t.Errorf("after Shift+Tab, expected SectionResults, got %v", model.focusedSection)
	}
}

func TestDashboardModel_HelpOverlay(t *testing.T) {
	cfg := config.AppConfig{N: 100}
	model := NewDashboardModel(cfg, newMockCalculators())
	model.ready = true
	model.width = 120
	model.height = 40

	// Initial state - no help overlay
	if model.helpOverlay {
		t.Fatal("expected helpOverlay=false initially")
	}

	// Press ? to toggle help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	result, _ := model.Update(msg)
	model = result.(DashboardModel)
	if !model.helpOverlay {
		t.Error("expected helpOverlay=true after pressing ?")
	}

	// Press Escape to close help
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	result, _ = model.Update(msg)
	model = result.(DashboardModel)
	if model.helpOverlay {
		t.Error("expected helpOverlay=false after pressing Escape")
	}
}

func TestDashboardModel_InputHandling(t *testing.T) {
	cfg := config.AppConfig{N: 100}
	model := NewDashboardModel(cfg, newMockCalculators())
	model.ready = true
	model.width = 120
	model.height = 40
	model.input.inputActive = true

	// Type a digit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}}
	result, _ := model.Update(msg)
	model = result.(DashboardModel)

	// N should now have 5 appended
	expected := "1005"
	if model.input.n != expected {
		t.Errorf("expected input.n='%s', got '%s'", expected, model.input.n)
	}

	// Backspace
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ = model.Update(msg)
	model = result.(DashboardModel)
	if model.input.n != "100" {
		t.Errorf("after backspace, expected '100', got '%s'", model.input.n)
	}
}

func TestSection_Next(t *testing.T) {
	tests := []struct {
		current  Section
		expected Section
	}{
		{SectionInput, SectionAlgorithms},
		{SectionAlgorithms, SectionResults},
		{SectionResults, SectionInput},
	}

	for _, tt := range tests {
		result := tt.current.Next()
		if result != tt.expected {
			t.Errorf("Section(%v).Next() = %v, expected %v", tt.current, result, tt.expected)
		}
	}
}

func TestSection_Prev(t *testing.T) {
	tests := []struct {
		current  Section
		expected Section
	}{
		{SectionInput, SectionResults},
		{SectionAlgorithms, SectionInput},
		{SectionResults, SectionAlgorithms},
	}

	for _, tt := range tests {
		result := tt.current.Prev()
		if result != tt.expected {
			t.Errorf("Section(%v).Prev() = %v, expected %v", tt.current, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Nanosecond, "500ns"},
		{1500 * time.Nanosecond, "1.5µs"},
		{1500 * time.Microsecond, "1.5ms"},
		{1500 * time.Millisecond, "1.50s"},
		{90 * time.Second, "1m30s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{100, "100"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{12345, "12,345"},
	}

	for _, tt := range tests {
		result := formatNumber(tt.n)
		if result != tt.expected {
			t.Errorf("formatNumber(%d) = %s, expected %s", tt.n, result, tt.expected)
		}
	}
}

func TestCheckConsistency(t *testing.T) {
	// Empty results
	if !checkConsistency(nil) {
		t.Error("expected empty results to be consistent")
	}

	// All same results - would need real big.Int for proper test
	// This is a basic structural test
}

func TestDashboardModel_View(t *testing.T) {
	cfg := config.AppConfig{N: 100}
	model := NewDashboardModel(cfg, newMockCalculators())
	model.ready = true
	model.width = 120
	model.height = 40

	// Should not panic and should return non-empty string
	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain key sections
	if !containsString(view, "FIBONACCI") {
		t.Error("view should contain 'FIBONACCI'")
	}
	if !containsString(view, "INPUT") {
		t.Error("view should contain 'INPUT'")
	}
	if !containsString(view, "ALGORITHMS") {
		t.Error("view should contain 'ALGORITHMS'")
	}
	if !containsString(view, "RESULT") {
		t.Error("view should contain 'RESULT'")
	}
}

func TestDashboardModel_HelpOverlayView(t *testing.T) {
	cfg := config.AppConfig{N: 100}
	model := NewDashboardModel(cfg, newMockCalculators())
	model.ready = true
	model.width = 120
	model.height = 40
	model.helpOverlay = true

	view := model.View()
	if !containsString(view, "HELP") {
		t.Error("help overlay should contain 'HELP'")
	}
	if !containsString(view, "Navigation") {
		t.Error("help overlay should contain 'Navigation'")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
