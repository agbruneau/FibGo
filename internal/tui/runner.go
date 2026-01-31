package tui

import (
	"context"
	"math/big"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/agbru/fibcalc/internal/fibonacci"
)

// CalcDoneMsg signals that one calculator has finished.
type CalcDoneMsg struct {
	CalcIndex int
	Name      string
	Result    *big.Int
	Duration  time.Duration
	Err       error
}

// tickMsg triggers periodic UI refresh during calculations.
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// runSingleCalc returns a tea.Cmd that runs one calculator.
// Progress updates are sent to the shared progressChan.
// The Cmd blocks until the calculator finishes and returns CalcDoneMsg.
func runSingleCalc(
	calc fibonacci.Calculator,
	idx int,
	n uint64,
	opts fibonacci.Options,
	ctx context.Context,
	progressChan chan<- fibonacci.ProgressUpdate,
) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		val, err := calc.Calculate(ctx, progressChan, idx, n, opts)
		return CalcDoneMsg{
			CalcIndex: idx,
			Name:      calc.Name(),
			Result:    val,
			Duration:  time.Since(start),
			Err:       err,
		}
	}
}
