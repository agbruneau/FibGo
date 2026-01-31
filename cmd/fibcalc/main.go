package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/agbru/fibcalc/internal/cli"
	apperrors "github.com/agbru/fibcalc/internal/errors"
	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/tui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Version information injected at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	os.Exit(run())
}

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	log.Logger = zerolog.Nop()
}

func run() int {
	// No arguments: launch the TUI
	if len(os.Args) == 1 {
		return runTUI()
	}

	n := flag.Uint64("n", 250_000_000, "Index of the Fibonacci number to calculate")
	algo := flag.String("algo", "all", "Algorithm to use: fast, matrix, fft, or all")
	timeout := flag.Duration("timeout", 5*time.Minute, "Timeout for the calculation")
	threshold := flag.Int("threshold", fibonacci.DefaultParallelThreshold, "Parallel threshold (bits)")
	fftThreshold := flag.Int("fft-threshold", fibonacci.DefaultFFTThreshold, "FFT threshold (bits)")
	verbose := flag.Bool("v", false, "Display the full (non-truncated) result")
	details := flag.Bool("d", false, "Display detailed execution metrics")
	concise := flag.Bool("c", false, "Display the calculated value")
	version := flag.Bool("version", false, "Display version information and exit")
	flag.Parse()

	if *version {
		fmt.Printf("fibcalc %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		return apperrors.ExitSuccess
	}

	out := os.Stdout

	cfg := cli.ExecutionConfig{
		N:            *n,
		Algo:         *algo,
		Timeout:      *timeout,
		Threshold:    *threshold,
		FFTThreshold: *fftThreshold,
	}

	factory := fibonacci.NewDefaultFactory()
	calculators := cli.GetCalculatorsToRun(cfg, factory)
	if len(calculators) == 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown algorithm %q. Available: fast, matrix, fft, all\n", *algo)
		return apperrors.ExitErrorConfig
	}

	cli.PrintExecutionConfig(cfg, out)
	cli.PrintExecutionMode(calculators, out)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	progressChan := make(chan fibonacci.ProgressUpdate, len(calculators)*100)
	var wg sync.WaitGroup
	wg.Add(1)
	go cli.DisplayProgress(&wg, progressChan, len(calculators), out)

	opts := fibonacci.Options{
		ParallelThreshold: cfg.Threshold,
		FFTThreshold:      cfg.FFTThreshold,
	}

	type calcResult struct {
		calc     fibonacci.Calculator
		val      *big.Int
		duration time.Duration
		err      error
	}

	results := make([]calcResult, len(calculators))
	var calcWg sync.WaitGroup

	for i, calc := range calculators {
		calcWg.Add(1)
		go func(idx int, c fibonacci.Calculator) {
			defer calcWg.Done()
			start := time.Now()
			val, err := c.Calculate(ctx, progressChan, idx, cfg.N, opts)
			results[idx] = calcResult{calc: c, val: val, duration: time.Since(start), err: err}
		}(i, calc)
	}

	calcWg.Wait()
	close(progressChan)
	wg.Wait()

	exitCode := apperrors.ExitSuccess
	colors := cli.CLIColorProvider{}

	sort.Slice(results, func(i, j int) bool {
		return results[i].duration < results[j].duration
	})

	fmt.Fprintf(out, "\n--- Results ---\n")
	for i, r := range results {
		if r.err != nil {
			code := apperrors.HandleCalculationError(r.err, r.duration, os.Stderr, colors)
			if code != apperrors.ExitSuccess {
				exitCode = code
			}
			continue
		}
		fmt.Fprintf(out, "=== %s (%s) ===\n", r.calc.Name(), cli.FormatExecutionDuration(r.duration))
		cli.DisplayResult(r.val, cfg.N, r.duration, *verbose, *details, *concise, out)
		if i < len(results)-1 {
			fmt.Fprintln(out)
		}
	}

	return exitCode
}

func runTUI() int {
	factory := fibonacci.NewDefaultFactory()
	m := tui.NewModel(factory, Version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return apperrors.ExitErrorGeneric
	}
	return apperrors.ExitSuccess
}
