package tui

import (
	"context"
	"math/big"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// AppState represents the current state of the TUI application.
type AppState int

const (
	StateConfig  AppState = iota // User is editing configuration
	StateRunning                 // Calculations are in progress
	StateResults                 // Calculations complete
)

// FocusField represents which configuration field has focus.
type FocusField int

const (
	FocusN       FocusField = iota // N input field
	FocusAlgo                      // Algorithm selector
	FocusTimeout                   // Timeout input field
	focusCount                     // sentinel for cycling
)

// AlgoProgress tracks the state of a single calculator during execution.
type AlgoProgress struct {
	Name     string
	Progress float64
	Done     bool
	Duration time.Duration
	Result   *big.Int
	Err      error
	BitLen   int
	Digits   int
}

// Model is the bubbletea model for the FibGo TUI.
type Model struct {
	state AppState
	focus FocusField

	// Configuration inputs
	nInput      textinput.Model
	algoIndex   int
	algoChoices []string
	timeoutInput textinput.Model

	// Fibonacci infrastructure
	factory fibonacci.CalculatorFactory

	// Calculation state
	algos        []AlgoProgress
	startTime    time.Time
	cancelFunc   context.CancelFunc
	doneCount    int
	progressChan chan fibonacci.ProgressUpdate

	// Layout
	width  int
	height int

	// App info
	version string

	// Status message
	statusMsg string
}

// NewModel creates a new TUI model with default configuration.
func NewModel(factory fibonacci.CalculatorFactory, version string) Model {
	ni := textinput.New()
	ni.Placeholder = "250000000"
	ni.SetValue("250000000")
	ni.Focus()
	ni.CharLimit = 20
	ni.Width = 20

	ti := textinput.New()
	ti.Placeholder = "5m"
	ti.SetValue("5m")
	ti.CharLimit = 10
	ti.Width = 10

	return Model{
		state:       StateConfig,
		focus:       FocusN,
		nInput:      ni,
		algoIndex:   0,
		algoChoices: []string{"all", "fast", "matrix", "fft"},
		timeoutInput: ti,
		factory:     factory,
		version:     version,
		statusMsg:   "Ready",
	}
}
