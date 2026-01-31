// Package cli provides the REPL (Read-Eval-Print Loop) functionality
// for interactive Fibonacci calculations.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agbru/fibcalc/internal/fibonacci"
	"github.com/agbru/fibcalc/internal/ui"
)

// REPLConfig holds configuration for the REPL session.
type REPLConfig struct {
	// DefaultAlgo is the default algorithm to use for calculations.
	DefaultAlgo string
	// Timeout is the maximum duration for each calculation.
	Timeout time.Duration
	// Threshold is the parallelism threshold.
	Threshold int
	// FFTThreshold is the FFT multiplication threshold.
	FFTThreshold int
	// HexOutput displays results in hexadecimal format.
	HexOutput bool
}

// REPL represents an interactive Fibonacci calculator session.
type REPL struct {
	config      REPLConfig
	registry    map[string]fibonacci.Calculator
	currentAlgo string
	in          io.Reader
	out         io.Writer
}

// NewREPL creates a new REPL instance.
//
// Parameters:
//   - registry: Map of available calculators.
//   - config: REPL configuration.
//
// Returns:
//   - *REPL: A new REPL instance.
func NewREPL(registry map[string]fibonacci.Calculator, config REPLConfig) *REPL {
	currentAlgo := config.DefaultAlgo
	if currentAlgo == "" || currentAlgo == "all" {
		// Pick the first available algorithm as default
		for name := range registry {
			currentAlgo = name
			break
		}
	}

	return &REPL{
		config:      config,
		registry:    registry,
		currentAlgo: currentAlgo,
		in:          os.Stdin,
		out:         os.Stdout,
	}
}

// SetInput sets a custom input reader (useful for testing).
func (r *REPL) SetInput(in io.Reader) {
	r.in = in
}

// SetOutput sets a custom output writer (useful for testing).
func (r *REPL) SetOutput(out io.Writer) {
	r.out = out
}

// Start begins the interactive REPL session.
// It continuously reads user input and processes commands until
// the user exits or EOF is reached.
func (r *REPL) Start() {
	r.printBanner()
	r.printHelp()
	fmt.Fprintln(r.out)

	reader := bufio.NewReader(r.in)

	for {
		fmt.Fprint(r.out, ui.ColorGreen()+"fib> "+ui.ColorReset())

		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(r.out, "\nGoodbye!")
				return
			}
			fmt.Fprintf(r.out, "%sRead error: %v%s\n", ui.ColorRed(), err, ui.ColorReset())
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if !r.processCommand(input) {
			return // Exit command received
		}
	}
}

// printBanner displays the REPL welcome banner.
func (r *REPL) printBanner() {
	fmt.Fprintf(r.out, "\n%s╔══════════════════════════════════════════════════════════╗%s\n", ui.ColorCyan(), ui.ColorReset())
	fmt.Fprintf(r.out, "%s║%s     %s🔢 Fibonacci Calculator - Interactive Mode%s            %s║%s\n",
		ui.ColorCyan(), ui.ColorReset(), ui.ColorBold(), ui.ColorReset(), ui.ColorCyan(), ui.ColorReset())
	fmt.Fprintf(r.out, "%s╚══════════════════════════════════════════════════════════╝%s\n\n", ui.ColorCyan(), ui.ColorReset())
}

// printHelp displays available commands.
func (r *REPL) printHelp() {
	fmt.Fprintf(r.out, "%sAvailable commands:%s\n", ui.ColorBold(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %scalc <n>%s      - Calculate F(n) with current algorithm\n", ui.ColorYellow(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %salgo <name>%s   - Change algorithm (%s)\n", ui.ColorYellow(), ui.ColorReset(), r.getAlgoList())
	fmt.Fprintf(r.out, "  %scompare <n>%s   - Compare all algorithms for F(n)\n", ui.ColorYellow(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %slist%s          - List available algorithms\n", ui.ColorYellow(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %shex%s           - Toggle hexadecimal display\n", ui.ColorYellow(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %sstatus%s        - Display current configuration\n", ui.ColorYellow(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %shelp%s          - Display this help\n", ui.ColorYellow(), ui.ColorReset())
	fmt.Fprintf(r.out, "  %sexit%s / %squit%s  - Exit interactive mode\n", ui.ColorYellow(), ui.ColorReset(), ui.ColorYellow(), ui.ColorReset())
}

// getAlgoList returns a comma-separated list of available algorithms.
func (r *REPL) getAlgoList() string {
	algos := make([]string, 0, len(r.registry))
	for name := range r.registry {
		algos = append(algos, name)
	}
	return strings.Join(algos, ", ")
}

// processCommand parses and executes a user command.
// Returns false if the REPL should exit.
func (r *REPL) processCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return true
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "calc", "c":
		r.cmdCalc(args)
	case "algo", "a":
		r.cmdAlgo(args)
	case "compare", "cmp":
		r.cmdCompare(args)
	case "list", "ls":
		r.cmdList()
	case "hex":
		r.cmdHex()
	case "status", "st":
		r.cmdStatus()
	case "help", "h", "?":
		r.printHelp()
	case "exit", "quit", "q":
		fmt.Fprintf(r.out, "%sGoodbye!%s\n", ui.ColorGreen(), ui.ColorReset())
		return false
	default:
		// Try to interpret as a number for quick calculation
		if n, err := strconv.ParseUint(cmd, 10, 64); err == nil {
			r.calculate(n)
		} else {
			fmt.Fprintf(r.out, "%sUnknown command: %s%s\n", ui.ColorRed(), cmd, ui.ColorReset())
			fmt.Fprintf(r.out, "Type %shelp%s to see available commands.\n", ui.ColorYellow(), ui.ColorReset())
		}
	}

	return true
}

// cmdCalc handles the "calc" command.
func (r *REPL) cmdCalc(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(r.out, "%sUsage: calc <n>%s\n", ui.ColorRed(), ui.ColorReset())
		return
	}

	n, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(r.out, "%sInvalid value: %s%s\n", ui.ColorRed(), args[0], ui.ColorReset())
		return
	}

	r.calculate(n)
}

// calculate performs a Fibonacci calculation with the current algorithm.
func (r *REPL) calculate(n uint64) {
	calc, ok := r.registry[r.currentAlgo]
	if !ok {
		fmt.Fprintf(r.out, "%sAlgorithm not found: %s%s\n", ui.ColorRed(), r.currentAlgo, ui.ColorReset())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	fmt.Fprintf(r.out, "Calculating F(%s%d%s) with %s%s%s...\n",
		ui.ColorMagenta(), n, ui.ColorReset(),
		ui.ColorCyan(), calc.Name(), ui.ColorReset())

	opts := fibonacci.Options{
		ParallelThreshold: r.config.Threshold,
		FFTThreshold:      r.config.FFTThreshold,
	}

	// Create a progress channel
	progressChan := make(chan fibonacci.ProgressUpdate, 10)

	// Use DisplayProgress to show a spinner and progress bar
	var wg sync.WaitGroup
	wg.Add(1)
	go DisplayProgress(&wg, progressChan, 1, r.out)

	start := time.Now()
	result, err := calc.Calculate(ctx, progressChan, 0, n, opts)
	duration := time.Since(start)
	close(progressChan)
	wg.Wait()

	if err != nil {
		fmt.Fprintf(r.out, "%sError: %v%s\n", ui.ColorRed(), err, ui.ColorReset())
		return
	}

	// Format duration
	durationStr := FormatExecutionDuration(duration)

	// Display result
	fmt.Fprintf(r.out, "\n%sResult:%s\n", ui.ColorBold(), ui.ColorReset())
	fmt.Fprintf(r.out, "  Time: %s%s%s\n", ui.ColorGreen(), durationStr, ui.ColorReset())
	fmt.Fprintf(r.out, "  Bits:  %s%d%s\n", ui.ColorCyan(), result.BitLen(), ui.ColorReset())

	resultStr := result.String()
	numDigits := len(resultStr)
	fmt.Fprintf(r.out, "  Digits: %s%d%s\n", ui.ColorCyan(), numDigits, ui.ColorReset())

	if r.config.HexOutput {
		hexStr := result.Text(16)
		if len(hexStr) > TruncationLimit {
			fmt.Fprintf(r.out, "  F(%d) = %s0x%s...%s%s (truncated)\n",
				n, ui.ColorGreen(), hexStr[:HexDisplayEdges], hexStr[len(hexStr)-HexDisplayEdges:], ui.ColorReset())
		} else {
			fmt.Fprintf(r.out, "  F(%d) = %s0x%s%s\n", n, ui.ColorGreen(), hexStr, ui.ColorReset())
		}
	} else if numDigits > TruncationLimit {
		fmt.Fprintf(r.out, "  F(%d) = %s%s...%s%s (truncated)\n",
			n, ui.ColorGreen(), resultStr[:DisplayEdges], resultStr[numDigits-DisplayEdges:], ui.ColorReset())
	} else {
		fmt.Fprintf(r.out, "  F(%d) = %s%s%s\n", n, ui.ColorGreen(), resultStr, ui.ColorReset())
	}
	fmt.Fprintln(r.out)
}

// cmdAlgo handles the "algo" command.
func (r *REPL) cmdAlgo(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(r.out, "%sUsage: algo <name>%s\n", ui.ColorRed(), ui.ColorReset())
		fmt.Fprintf(r.out, "Available algorithms: %s\n", r.getAlgoList())
		return
	}

	name := strings.ToLower(args[0])
	if _, ok := r.registry[name]; !ok {
		fmt.Fprintf(r.out, "%sUnknown algorithm: %s%s\n", ui.ColorRed(), name, ui.ColorReset())
		fmt.Fprintf(r.out, "Available algorithms: %s\n", r.getAlgoList())
		return
	}

	r.currentAlgo = name
	fmt.Fprintf(r.out, "Algorithm changed to: %s%s%s\n", ui.ColorGreen(), r.registry[name].Name(), ui.ColorReset())
}

// cmdCompare handles the "compare" command.
func (r *REPL) cmdCompare(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(r.out, "%sUsage: compare <n>%s\n", ui.ColorRed(), ui.ColorReset())
		return
	}

	n, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(r.out, "%sInvalid value: %s%s\n", ui.ColorRed(), args[0], ui.ColorReset())
		return
	}

	fmt.Fprintf(r.out, "\n%sComparison for F(%d):%s\n", ui.ColorBold(), n, ui.ColorReset())
	fmt.Fprintf(r.out, "%s─────────────────────────────────────────────%s\n", ui.ColorCyan(), ui.ColorReset())

	opts := fibonacci.Options{
		ParallelThreshold: r.config.Threshold,
		FFTThreshold:      r.config.FFTThreshold,
	}

	results := make(map[string]string)
	var firstResult string

	for name, calc := range r.registry {
		ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)

		// Create a progress channel for this calculation
		progressChan := make(chan fibonacci.ProgressUpdate, 10)
		go func() {
			for range progressChan {
				// Discard progress updates
			}
		}()

		start := time.Now()
		result, err := calc.Calculate(ctx, progressChan, 0, n, opts)
		duration := time.Since(start)
		close(progressChan)
		cancel()

		if err != nil {
			fmt.Fprintf(r.out, "  %s%-20s%s: %sError - %v%s\n",
				ui.ColorYellow(), name, ui.ColorReset(),
				ui.ColorRed(), err, ui.ColorReset())
			continue
		}

		durationStr := FormatExecutionDuration(duration)
		resultStr := result.String()
		results[name] = resultStr

		if firstResult == "" {
			firstResult = resultStr
		}

		// Check consistency
		status := ui.ColorGreen() + "✓" + ui.ColorReset()
		if resultStr != firstResult {
			status = ui.ColorRed() + "✗ INCONSISTENT" + ui.ColorReset()
		}

		fmt.Fprintf(r.out, "  %s%-20s%s: %s%12s%s %s\n",
			ui.ColorYellow(), name, ui.ColorReset(),
			ui.ColorCyan(), durationStr, ui.ColorReset(),
			status)
	}

	fmt.Fprintf(r.out, "%s─────────────────────────────────────────────%s\n\n", ui.ColorCyan(), ui.ColorReset())
}

// cmdList handles the "list" command.
func (r *REPL) cmdList() {
	fmt.Fprintf(r.out, "\n%sAvailable algorithms:%s\n", ui.ColorBold(), ui.ColorReset())
	for name, calc := range r.registry {
		marker := "  "
		if name == r.currentAlgo {
			marker = ui.ColorGreen() + "► " + ui.ColorReset()
		}
		fmt.Fprintf(r.out, "%s%s%-10s%s - %s\n", marker, ui.ColorYellow(), name, ui.ColorReset(), calc.Name())
	}
	fmt.Fprintln(r.out)
}

// cmdHex toggles hexadecimal output mode.
func (r *REPL) cmdHex() {
	r.config.HexOutput = !r.config.HexOutput
	status := "disabled"
	if r.config.HexOutput {
		status = "enabled"
	}
	fmt.Fprintf(r.out, "Hexadecimal display: %s%s%s\n", ui.ColorGreen(), status, ui.ColorReset())
}

// cmdStatus displays current REPL configuration.
func (r *REPL) cmdStatus() {
	fmt.Fprintf(r.out, "\n%sCurrent configuration:%s\n", ui.ColorBold(), ui.ColorReset())
	fmt.Fprintf(r.out, "  Algorithm:      %s%s%s\n", ui.ColorCyan(), r.currentAlgo, ui.ColorReset())
	fmt.Fprintf(r.out, "  Timeout:        %s%s%s\n", ui.ColorCyan(), r.config.Timeout, ui.ColorReset())
	fmt.Fprintf(r.out, "  Threshold:      %s%d%s bits\n", ui.ColorCyan(), r.config.Threshold, ui.ColorReset())
	fmt.Fprintf(r.out, "  FFT Threshold:  %s%d%s bits\n", ui.ColorCyan(), r.config.FFTThreshold, ui.ColorReset())
	hexStatus := "no"
	if r.config.HexOutput {
		hexStatus = "yes"
	}
	fmt.Fprintf(r.out, "  Hexadecimal:    %s%s%s\n", ui.ColorCyan(), hexStatus, ui.ColorReset())
	fmt.Fprintln(r.out)
}
