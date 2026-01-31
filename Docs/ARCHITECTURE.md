# Fibonacci Calculator Architecture

> **Version**: 1.3.0
> **Last Updated**: January 2026

## Overview

The Fibonacci Calculator is designed according to **Clean Architecture** principles, with strict separation of responsibilities and low coupling between modules. This architecture enables maximum testability, easy scalability, and simplified maintenance.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           ENTRY POINTS                                  │
│                                                                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │
│  │   CLI   │  │ Server  │  │ Docker  │  │  REPL   │  │   TUI   │       │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘       │
│       │            │            │            │            │            │
│       └────────────┼────────────┼────────────┼────────────┘            │
│                    ▼            ▼            ▼                         │
│              ┌───────────────┐ ┌────────────────┐ ┌────────────────┐   │
│              │ cmd/fibcalc   │ │ internal/cli   │ │ internal/tui   │   │
│              │   main.go     │ │   repl.go      │ │   tui.go       │   │
│              └───────┬───────┘ └───────┬────────┘ └───────┬────────┘   │
└──────────────────────┼─────────────────┼──────────────────┼─────────────┘
                       │                 │                  │
                       └─────────────────┼──────────────────┘
                                         │
┌─────────────────────────────────────┼───────────────────────────────────┐
│                   ORCHESTRATION LAYER                                   │
│                                     ▼                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    internal/orchestration                        │   │
│  │  • ExecuteCalculations() - Parallel algorithm execution         │   │
│  │  • AnalyzeComparisonResults() - Analysis and comparison         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                            │                                           │
│  ┌─────────────────────────┼───────────────────────────────────────┐   │
│  │                         ▼                                        │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │   │
│  │  │   config    │  │ calibration │  │   server    │              │   │
│  │  │   Parsing   │  │   Tuning    │  │   HTTP API  │              │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└────────────────────────────┼────────────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────────┐
│                      BUSINESS LAYER                                     │
│                            ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    internal/fibonacci                            │   │
│  │                                                                  │   │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────┐ │   │
│  │  │  Fast Doubling   │  │     Matrix       │  │    FFT-Based   │ │   │
│  │  │  O(log n)        │  │  Exponentiation  │  │    Doubling    │ │   │
│  │  │  Parallel        │  │  O(log n)        │  │    O(log n)    │ │   │
│  │  │  Zero-Alloc      │  │  Strassen        │  │    FFT Mul     │ │   │
│  │  └──────────────────┘  └──────────────────┘  └────────────────┘ │   │
│  │                            │                                     │   │
│  │                            ▼                                     │   │
│  │  ┌─────────────────────────────────────────────────────────────┐│   │
│  │  │                    internal/bigfft                          ││   │
│  │  │  • FFT multiplication for very large numbers                ││   │
│  │  │  • Complexity O(n log n) vs O(n^1.585) for Karatsuba        ││   │
│  │  └─────────────────────────────────────────────────────────────┘│   │
│  └─────────────────────────────────────────────────────────────────┘   │
└────────────────────────────┼────────────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────────┐
│                   PRESENTATION LAYER                                    │
│                            ▼                                           │
│  ┌──────────────────────────────────┐  ┌────────────────────────────┐  │
│  │         internal/cli             │  │       internal/tui         │  │
│  │  • Spinner and progress bar      │  │  • Elm Architecture        │  │
│  │  • Result formatting             │  │  • Navigation and views    │  │
│  │  • Colour themes                 │  │  • Real-time progress      │  │
│  │  • NO_COLOR support              │  │  • Theme integration       │  │
│  └──────────────────────────────────┘  └────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Package Structure

### `cmd/fibcalc`

Application entry point. Responsibilities:

- Command-line argument parsing
- Component initialization
- Routing to CLI or server mode
- System signal handling

### `internal/fibonacci`

Business core of the application. Contains:

- **`calculator.go`**: `Calculator` interface and generic wrapper
- **`fastdoubling.go`**: Optimized Fast Doubling algorithm
- **`matrix.go`**: Matrix exponentiation with Strassen
- **`fft_based.go`**: Calculator forcing FFT multiplication
- **`fft.go`**: Multiplication selection logic (standard vs FFT)
- **`constants.go`**: Thresholds and configuration constants

### `internal/bigfft`

FFT multiplication implementation for `big.Int`:

- **`fft.go`**: Main FFT algorithm
- **`fermat.go`**: Modular arithmetic for FFT
- **`pool.go`**: Object pools to reduce allocations

### `internal/orchestration`

Concurrent execution management with Clean Architecture decoupling:

- Parallel execution of multiple algorithms
- Result aggregation and comparison
- Error and timeout handling
- **`ProgressReporter` interface**: Decouples progress display from orchestration logic
- **`ResultPresenter` interface**: Decouples result presentation from analysis logic
- **`NullProgressReporter`**: No-op implementation for quiet mode and testing

### `internal/calibration`

Automatic calibration system:

- Optimal threshold detection for the hardware
- Calibration profile persistence
- Adaptive threshold generation based on CPU

### `internal/server`

HTTP REST server:

- `/calculate`, `/health`, `/algorithms`, `/metrics` endpoints
- Rate limiting and security
- Logging and metrics middleware
- Graceful shutdown

### `internal/cli`

Command-line user interface:

- Animated spinner with progress bar
- Estimated time remaining (ETA)
- Colour theme system (dark, light, none)
- Large number formatting
- **REPL Mode** (`repl.go`): Interactive session for multiple calculations
  - Commands: `calc`, `algo`, `compare`, `list`, `hex`, `status`, `help`, `exit`
  - On-the-fly algorithm switching
  - Real-time algorithm comparison
- Autocompletion script generation (bash, zsh, fish, powershell)
- `NO_COLOR` environment variable support
- **Interface Implementations** (`presenter.go`):
  - `CLIProgressReporter`: Implements `orchestration.ProgressReporter` for CLI progress display
  - `CLIResultPresenter`: Implements `orchestration.ResultPresenter` for CLI result formatting

### `internal/tui`

Rich Terminal User Interface using the Charm stack (Bubbletea, Bubbles, Lipgloss).
Redesigned as a single-screen HTOP-style dashboard:

- **`tui.go`**: Entry point, `tea.NewProgram()` initialization
- **`dashboard.go`**: DashboardModel with consolidated state (Init, Update, View)
- **`sections.go`**: Section type (Input, Algorithms, Results) and navigation helpers
- **Dashboard Sections**:
  - `dashboard_input.go`: Input section (N field, calculate/compare buttons with left/right navigation and button focus tracking via `buttonIndex`)
  - `dashboard_algorithms.go`: Algorithm table with real-time progress bars and adaptive separator width
  - `dashboard_results.go`: Results display section with detail toggle
  - `dashboard_overlays.go`: Help overlay and responsive header layout
- **`messages.go`**: Message types for state updates (ProgressMsg, ResultMsg, etc.)
- **`commands.go`**: Async commands for calculations and progress listening
- **`keys.go`**: Keyboard bindings (section navigation, actions, quit)
- **`styles.go`**: Lipgloss styles integrated with `internal/ui` themes
- **`presenter.go`**: Interface implementations:
  - `TUIProgressReporter`: Bridges orchestration progress to Bubbletea messages
  - `TUIResultPresenter`: No-op (TUI handles results via messages)
- **`model.go`**: Legacy model kept for backward compatibility

### `internal/config`

Configuration management:

- CLI flag parsing
- Parameter validation
- Default values

### `internal/errors`

Centralised error handling:

- Custom error types
- Standardised exit codes

## Architecture Decision Records (ADR)

### ADR-001: Using `sync.Pool` for Calculation States

**Context**: Fibonacci calculations for large N require numerous temporary `big.Int` objects.

**Decision**: Use `sync.Pool` to recycle calculation states (`calculationState`, `matrixState`).

**Consequences**:

- ✅ Drastic reduction in memory allocations
- ✅ Decreased GC pressure
- ✅ 20-30% performance improvement
- ⚠️ Increased code complexity

### ADR-002: Dynamic Multiplication Algorithm Selection

**Context**: FFT multiplication is more efficient than Karatsuba for very large numbers, but has significant overhead for small numbers.

**Decision**: Implement a `smartMultiply` function that selects the algorithm based on operand size.

**Consequences**:

- ✅ Optimal performance across the entire value range
- ✅ Configurable via `--fft-threshold`
- ⚠️ Requires calibration for each architecture

### ADR-003: Hexagonal Architecture for the Server

**Context**: The server must be testable and extensible.

**Decision**: Use interfaces and dependency injection via functional options.

**Consequences**:

- ✅ Facilitated unit testing
- ✅ Easily composable middleware
- ✅ Flexible configuration

### ADR-004: Adaptive Parallelism

**Context**: Parallelism has a synchronization cost that can exceed gains for small calculations.

**Decision**: Enable parallelism only above a configurable threshold (`--threshold`).

**Consequences**:

- ✅ Optimal performance according to calculation size
- ✅ Avoids CPU saturation for small N
- ⚠️ Parallelism disabled when FFT is used (FFT already saturates CPU)

### ADR-005: Interface-Based Decoupling (Orchestration → CLI)

**Context**: The orchestration package was directly importing CLI packages, violating Clean Architecture principles where business logic should not depend on presentation.

**Decision**: Define `ProgressReporter` and `ResultPresenter` interfaces in the orchestration package, with implementations in the CLI package.

**Consequences**:

- ✅ Clean Architecture compliance: orchestration no longer imports CLI
- ✅ Improved testability: interfaces can be mocked for unit tests
- ✅ Flexibility: alternative presenters (JSON, GUI) can be easily added
- ✅ `NullProgressReporter` enables quiet mode without conditionals
- ⚠️ Slightly more complex initialization in the app layer

## Data Flow

### CLI Mode

```
1. app.New() parses arguments → config.AppConfig
2. app.Run() dispatches to appropriate mode
3. If --calibrate: calibration.RunCalibration() and exit
4. If --auto-calibrate: calibration.AutoCalibrate() updates config
5. cli.GetCalculatorsToRun() selects algorithms
6. orchestration.ExecuteCalculations() launches parallel calculations
   - Each Calculator.Calculate() executes in a goroutine
   - Progress updates are sent on a channel
   - ProgressReporter (CLIProgressReporter) displays progress
7. orchestration.AnalyzeComparisonResults() analyzes results
   - ResultPresenter (CLIResultPresenter) formats and displays output
```

### Server Mode

```
1. main() detects --server and calls server.NewServer()
2. Server.Start() starts HTTP server with graceful shutdown
3. For each /calculate request:
   a. SecurityMiddleware checks headers
   b. RateLimitMiddleware applies rate limiting
   c. loggingMiddleware logs the request
   d. metricsMiddleware records metrics
   e. handleCalculate() executes the calculation
4. Result is returned as JSON
```

### Interactive Mode (REPL)

```
1. main() detects --interactive and calls cli.NewREPL()
2. REPL.Start() displays banner and help
3. Main loop:
   a. Displays "fib> " prompt
   b. Reads user input
   c. Parses and executes command:
      - calc <n>: Calculation with current algorithm
      - algo <name>: Changes active algorithm
      - compare <n>: Compares all algorithms
      - list: Lists algorithms
      - hex: Toggles hexadecimal format
      - status: Displays configuration
      - exit: Ends session
4. Repeats until exit or EOF
```

### TUI Mode (HTOP-style Dashboard)

```
1. main() detects --tui and calls app.runTUI()
2. tui.Run() initializes tea.Program with DashboardModel
3. DashboardModel.Init() returns initial state (single-screen dashboard)
4. Bubbletea event loop:
   a. Update() receives messages (key events, progress updates, results)
   b. Global shortcuts handled first (Tab, Escape, Help)
   c. Section-specific handlers update focused section
   d. View() renders all sections on single screen
5. Dashboard sections (all visible at once):
   - Input: N field and action buttons (Calculate, Compare)
     * Left/Right navigation between input field and buttons
     * buttonIndex tracks which button is focused (0=Calculate, 1=Compare)
   - Algorithms: Table with real-time progress bars for all algorithms
   - Results: Calculation results with formatting options
6. Calculation flow:
   a. User enters N in input section, presses Enter or 'c'
   b. If on Calculate button: startSingleCalculation() triggers runCalculation
   c. If on Compare button: startComparison() runs all algorithms
   d. ProgressMsg updates progress bar for running algorithm(s)
   e. CalculationResultMsg/ComparisonResultsMsg updates results section
7. Comparison flow:
   a. User presses 'm' or Enter on Compare button
   b. startComparison() runs all calculators in parallel
   c. All progress bars update simultaneously
   d. ComparisonResultsMsg updates algorithms table and results
8. Program exits on 'q' or Ctrl+C
```

#### TUI Architecture Pattern (HTOP-style Dashboard)

```
┌─────────────────────────────────────────────────────────┐
│                    Bubbletea Runtime                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   ┌─────────┐    ┌─────────┐    ┌─────────┐           │
│   │  Init   │───▶│ Update  │───▶│  View   │           │
│   └─────────┘    └────┬────┘    └────┬────┘           │
│                       │              │                 │
│                       │              ▼                 │
│                       │    ┌─────────────────────┐    │
│                       │    │  Single Dashboard   │    │
│                       │    │  ┌───────────────┐  │    │
│                       │    │  │ Input Section │  │    │
│                       │    │  ├───────────────┤  │    │
│                       │    │  │ Algorithms    │  │    │
│                       │    │  │ (Progress)    │  │    │
│                       │    │  ├───────────────┤  │    │
│                       │    │  │ Results       │  │    │
│                       │    │  └───────────────┘  │    │
│                       │    └─────────────────────┘    │
│                       ▼                               │
│   ┌─────────────────────────────────────────┐        │
│   │               Commands                   │        │
│   │  • listenForProgress() - Channel bridge │        │
│   │  • runCalculation() - Async calculation │        │
│   │  • runComparison() - All algorithms     │        │
│   │  • tickCmd() - Animation ticks          │        │
│   └─────────────────────────────────────────┘        │
│                       │                               │
│                       ▼                               │
│   ┌─────────────────────────────────────────┐        │
│   │               Messages                   │        │
│   │  • ProgressMsg (with CalculatorIndex)   │        │
│   │  • CalculationResultMsg                 │        │
│   │  • ComparisonResultsMsg                 │        │
│   │  • KeyMsg, ThemeChangedMsg              │        │
│   └─────────────────────────────────────────┘        │
│                                                       │
└───────────────────────────────────────────────────────┘
```

## Performance Considerations

1. **Zero-Allocation**: Object pools avoid allocations in critical loops
2. **Smart Parallelism**: Enabled only when beneficial
3. **Adaptive FFT**: Used for very large numbers only
4. **Strassen**: Enabled for matrices with large elements
5. **Symmetric Squaring**: Specific optimization reducing multiplications

## Extensibility

To add a new algorithm:

1. Create a structure implementing the `coreCalculator` interface in `internal/fibonacci`
2. Register the calculator in `calculatorRegistry` in `main.go`
3. Add corresponding tests

To add a new API endpoint:

1. Add the handler in `internal/server/server.go`
2. Register the route in `NewServer()`
3. Update the OpenAPI documentation
