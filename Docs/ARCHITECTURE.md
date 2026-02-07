# Fibonacci Calculator Architecture

## Overview

The Fibonacci Calculator is designed according to **Clean Architecture** principles, with strict separation of responsibilities and low coupling between modules. This architecture enables maximum testability, easy scalability, and simplified maintenance.

**Go Module**: `github.com/agbru/fibcalc` (Go 1.25.0)

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           ENTRY POINT                                   │
│                                                                         │
│                    ┌────────────────────────┐                           │
│                    │  cmd/fibcalc            │                          │
│                    │  cmd/generate-golden    │                          │
│                    └───────────┬─────────────┘                          │
└────────────────────────────────┼─────────────────────────────────────────┘
                                 │
┌────────────────────────────────┼─────────────────────────────────────────┐
│                   ORCHESTRATION LAYER                                    │
│                                ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    internal/orchestration                         │   │
│  │  • ExecuteCalculations() — parallel algorithm execution           │   │
│  │  • AnalyzeComparisonResults() — analysis and comparison           │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                             │                                            │
│  ┌──────────────┐  ┌───────────────┐  ┌──────────────┐                  │
│  │    config     │  │  calibration  │  │     app      │                  │
│  │   Parsing     │  │    Tuning     │  │  Lifecycle   │                  │
│  └──────────────┘  └───────────────┘  └──────────────┘                  │
└────────────────────────────────┼─────────────────────────────────────────┘
                                 │
┌────────────────────────────────┼─────────────────────────────────────────┐
│                      BUSINESS LAYER                                      │
│                                ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    internal/fibonacci                             │   │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────┐ │   │
│  │  │  Fast Doubling   │  │     Matrix       │  │    FFT-Based   │ │   │
│  │  │  O(log n)        │  │  Exponentiation  │  │    Doubling    │ │   │
│  │  │  Parallel        │  │  O(log n)        │  │    O(log n)    │ │   │
│  │  │  Zero-Alloc      │  │  Strassen        │  │    FFT Mul     │ │   │
│  │  └──────────────────┘  └──────────────────┘  └────────────────┘ │   │
│  │                            │                                     │   │
│  │                            ▼                                     │   │
│  │  ┌──────────────────────────────────────────────────────────────┐│   │
│  │  │                    internal/bigfft                           ││   │
│  │  │  • FFT multiplication for very large numbers                ││   │
│  │  │  • Object pooling, bump allocation, SIMD dispatch           ││   │
│  │  └──────────────────────────────────────────────────────────────┘│   │
│  └──────────────────────────────────────────────────────────────────┘   │
└────────────────────────────────┼─────────────────────────────────────────┘
                                 │
┌────────────────────────────────┼─────────────────────────────────────────┐
│                   PRESENTATION LAYER                                     │
│                                ▼                                         │
│  ┌──────────────────────────────────┐  ┌──────────────────────────────┐ │
│  │         internal/cli             │  │       internal/ui            │ │
│  │  • Spinner and progress bar      │  │  • ANSI color functions     │ │
│  │  • Result formatting             │  │  • Theme system             │ │
│  │  • ETA estimation                │  │  • NO_COLOR support         │ │
│  │  • Shell completion              │  │                              │ │
│  └──────────────────────────────────┘  └──────────────────────────────┘ │
│  ┌──────────────────────────────────┐                                   │
│  │         internal/tui             │  Activated via --tui flag         │
│  │  • btop-style dashboard          │                                   │
│  │  • Real-time logs, metrics       │                                   │
│  │  • Progress bar with ETA         │                                   │
│  │  • Keyboard navigation           │                                   │
│  └──────────────────────────────────┘                                   │
└──────────────────────────────────────────────────────────────────────────┘
```

## Package Structure

### `cmd/fibcalc`

The main CLI entry point. A minimal wrapper that calls `app.New()` and `app.Run()`, providing:

- Version flag handling (`--version`)
- Command-line argument parsing via `internal/config`
- Application lifecycle (timeout + signal handling)

### `cmd/generate-golden`

Utility for generating Fibonacci golden test data used by the test suite.

- **`main.go`**: Entry point for golden file generation

### `internal/fibonacci`

Business core of the application. Contains algorithm implementations, the factory/registry system, multiplication strategies, and the observer pattern for progress reporting.

| File | Responsibility |
|------|---------------|
| `calculator.go` | `Calculator` and `coreCalculator` interfaces, `FibCalculator` decorator |
| `registry.go` | `CalculatorFactory` interface, `DefaultFactory` with lazy creation and caching |
| `strategy.go` | `Multiplier` (narrow) and `DoublingStepExecutor` (wide) interfaces, `MultiplicationStrategy` (deprecated alias); `AdaptiveStrategy`, `FFTOnlyStrategy`, `KaratsubaStrategy` |
| `observer.go` | `ProgressObserver` interface, `ProgressSubject` (observable) |
| `observers.go` | Observer implementations: `ChannelObserver`, `LoggingObserver`, `NoOpObserver` |
| `options.go` | `Options` struct: `ParallelThreshold`, `FFTThreshold`, `StrassenThreshold`, FFT cache settings (`FFTCacheMinBitLen`, `FFTCacheMaxEntries`, `FFTCacheEnabled`), dynamic threshold settings (`EnableDynamicThresholds`, `DynamicAdjustmentInterval`); `normalizeOptions()` fills zero values with defaults |
| `constants.go` | Performance tuning constants: `DefaultParallelThreshold` (4096), `DefaultFFTThreshold` (500,000), `DefaultStrassenThreshold` (3072), `ParallelFFTThreshold` (5,000,000), `CalibrationN` (10,000,000), `ProgressReportThreshold` (0.01) |
| `threshold_types.go` | Threshold type definitions |
| `dynamic_threshold.go` | Runtime threshold adjustment logic |
| `fastdoubling.go` | `OptimizedFastDoubling` algorithm implementation, `CalculationState` type and pool |
| `doubling_framework.go` | `DoublingFramework` — shared iteration framework for doubling-based algorithms |
| `matrix.go` | `MatrixExponentiation` algorithm implementation |
| `matrix_framework.go` | `MatrixFramework` — shared framework for matrix-based algorithms |
| `matrix_ops.go` | Matrix multiplication and squaring operations, Strassen dispatch (`multiplyMatrices`, `multiplyMatrixStrassen`), runtime threshold control (`Set/GetDefaultStrassenThreshold`) |
| `matrix_types.go` | `matrix` type (2x2), `matrixState` pool type |
| `fft_based.go` | `FFTBasedCalculator` — forces FFT for all multiplications |
| `fft.go` | `smartMultiply` / `smartSquare` — 2-tier multiplication selection (FFT or standard math/big) |
| `progress.go` | `ProgressCallback` type, progress utilities (`CalcTotalWork`, `ReportStepProgress`) |
| `common.go` | Task semaphore, `MaxPooledBitLen`, `executeTasks` generics, `executeMixedTasks` |
| `generator.go` | `SequenceGenerator` interface for Fibonacci sequence generation |
| `generator_iterative.go` | Iterative generator implementation |
| `testing.go` | Test helpers and utilities |
| `calculator_gmp.go` | GMP calculator, auto-registers via `init()` (build tag: `gmp`) |

### `internal/bigfft`

FFT multiplication for `big.Int`, with object pooling, memory management, and platform-specific SIMD dispatch.

| File | Responsibility |
|------|---------------|
| `fft.go` | Public API: `Mul`, `MulTo`, `Sqr`, `SqrTo` |
| `fft_core.go` | Core FFT algorithm implementation |
| `fft_recursion.go` | Recursive FFT decomposition |
| `fft_poly.go` | Polynomial operations for FFT |
| `fft_cache.go` | FFT transform caching |
| `fermat.go` | Fermat ring arithmetic: `fermat` type (Z/(2^k+1)), `Shift`, `ShiftHalf`, `Add`, `Sub`, `Mul`, `Sqr`, `norm`; `smallMulThreshold` for schoolbook/big.Int cutover |
| `pool.go` | `sync.Pool`-based object pools with size classes |
| `pool_warming.go` | Pool pre-warming for adaptive buffer pre-allocation |
| `allocator.go` | Memory allocator abstraction |
| `bump.go` | Bump allocator for batch allocations |
| `memory_est.go` | Memory estimation for pre-allocation |
| `scan.go` | Bit scanning utilities |
| `fft_recursion.go` | Recursive FFT decomposition with runtime-configurable parallelism (`FFTParallelismConfig`, `Set/GetFFTParallelismConfig`) |
| `arith_amd64.go` | amd64 vector arithmetic wrappers |
| `arith_generic.go` | Non-amd64 vector arithmetic wrappers |
| `arith_decl.go` | `go:linkname` declarations to `math/big` internals |
| `cpu_amd64.go` | Runtime CPU feature detection |

### `internal/orchestration`

Concurrent execution management with Clean Architecture decoupling.

| File | Responsibility |
|------|---------------|
| `orchestrator.go` | `ExecuteCalculations()`, `AnalyzeComparisonResults()` — parallel execution via `errgroup` |
| `interfaces.go` | `ProgressReporter`, `ResultPresenter` interfaces, `NullProgressReporter` |
| `calculator_selection.go` | `GetCalculatorsToRun()` — calculator selection logic from config |

### `internal/cli`

Command-line user interface and presentation layer.

| File | Responsibility |
|------|---------------|
| `output.go` | `Display*` / `Format*` / `Write*` functions for output |
| `presenter.go` | `CLIProgressReporter` and `CLIResultPresenter` implementations |
| `ui.go` | Spinner management and terminal interaction |
| `ui_display.go` | Display functions for progress reporting and result presentation |
| `ui_format.go` | Number formatting and duration/ETA formatting utilities |
| `progress_eta.go` | ETA estimation algorithm |
| `calculate.go` | Calculation orchestration entry point for CLI |
| `completion.go` | Shell completion script generation (bash, zsh, fish, powershell) |
| `provider.go` | Dependency provider for CLI components |

### `internal/tui`

Interactive TUI dashboard (btop-style), activated via `--tui` flag or `FIBCALC_TUI=true`. Built on the Elm architecture (Model-Update-View) with Bubble Tea.

| File | Responsibility |
|------|---------------|
| `doc.go` | Package documentation |
| `messages.go` | Tea message types (`ProgressMsg`, `ResultMsg`, `TickMsg`, `MemStatsMsg`, etc.) |
| `styles.go` | btop-inspired dark theme palette with lipgloss (rounded borders, color scheme) |
| `keymap.go` | Keyboard bindings (`q`, `space`, `r`, arrows, `pgup`/`pgdn`) |
| `bridge.go` | `TUIProgressReporter` and `TUIResultPresenter` — implements orchestration interfaces |
| `header.go` | Header sub-model (title, version, elapsed time using `FormatExecutionDuration`) |
| `logs.go` | Scrollable log panel sub-model (viewport, auto-scroll) |
| `metrics.go` | Runtime metrics sub-model (memory, heap, GC, goroutines, speed) |
| `chart.go` | Progress bar and ETA display sub-model |
| `sparkline.go` | Sparkline visualization for progress chart |
| `footer.go` | Footer sub-model (keyboard shortcuts, status indicator) |
| `model.go` | Root model, `Init()`/`Update()`/`View()`, `Run()` entry point, layout (60/40 split) |

### `internal/calibration`

Automatic calibration system for hardware-specific threshold tuning.

| File | Responsibility |
|------|---------------|
| `calibration.go` | Core calibration logic |
| `adaptive.go` | Adaptive threshold generation based on CPU |
| `microbench.go` | Micro-benchmarking routines |
| `io.go` | Calibration profile I/O |
| `profile.go` | Calibration profile data structures |
| `runner.go` | Calibration test runner |

### `internal/config`

Configuration management.

| File | Responsibility |
|------|---------------|
| `config.go` | `ParseConfig()`, `AppConfig` struct, flag parsing |
| `env.go` | Environment variable support (`FIBCALC_*` prefix) |
| `usage.go` | Help text and usage formatting |

### `internal/errors`

Centralized error handling.

| File | Responsibility |
|------|---------------|
| `errors.go` | Custom error types: `ConfigError`, `CalculationError` |
| `handler.go` | Error handler with standardized exit codes (0=success, 1=generic, 2=timeout, 3=mismatch, 4=config, 130=canceled) |

### `internal/app`

Application lifecycle management.

| File | Responsibility |
|------|---------------|
| `app.go` | Application initialization and lifecycle (`SetupContext`, signal handling) |
| `version.go` | Version information |
| `doc.go` | Package documentation |

### `internal/ui`

Terminal UI utilities.

| File | Responsibility |
|------|---------------|
| `colors.go` | ANSI color functions |
| `themes.go` | Theme system (dark, light, none), `NO_COLOR` support |

## Key Interfaces

### Calculator (public)

```go
type Calculator interface {
    Calculate(ctx context.Context, progressChan chan<- ProgressUpdate,
        calcIndex int, n uint64, opts Options) (*big.Int, error)
    Name() string
}
```

### coreCalculator (internal)

```go
type coreCalculator interface {
    CalculateCore(ctx context.Context, reporter ProgressCallback,
        n uint64, opts Options) (*big.Int, error)
    Name() string
}
```

### CalculatorFactory

```go
type CalculatorFactory interface {
    Create(name string) (Calculator, error)
    Get(name string) (Calculator, error)
    List() []string
    Register(name string, creator func() coreCalculator) error
    GetAll() map[string]Calculator
}
```

### Multiplier (narrow)

```go
type Multiplier interface {
    Multiply(z, x, y *big.Int, opts Options) (*big.Int, error)
    Square(z, x *big.Int, opts Options) (*big.Int, error)
    Name() string
}
```

### DoublingStepExecutor (wide)

```go
type DoublingStepExecutor interface {
    Multiplier
    ExecuteStep(ctx context.Context, s *CalculationState, opts Options, inParallel bool) error
}

// MultiplicationStrategy is a deprecated type alias for DoublingStepExecutor.
type MultiplicationStrategy = DoublingStepExecutor
```

### ProgressObserver

```go
type ProgressObserver interface {
    Update(calcIndex int, progress float64)
}
```

### ProgressReporter (orchestration)

```go
type ProgressReporter interface {
    DisplayProgress(wg *sync.WaitGroup, progressChan <-chan fibonacci.ProgressUpdate,
        numCalculators int, out io.Writer)
}
```

### ResultPresenter (orchestration)

```go
type ResultPresenter interface {
    PresentComparisonTable(results []CalculationResult, out io.Writer)
    PresentResult(result CalculationResult, n uint64, verbose, details, concise bool, out io.Writer)
    FormatDuration(d time.Duration) string
    HandleError(err error, duration time.Duration, out io.Writer) int
}
```

## Architecture Decision Records (ADR)

### ADR-001: Using `sync.Pool` for Calculation States

**Context**: Fibonacci calculations for large N require numerous temporary `big.Int` objects.

**Decision**: Use `sync.Pool` to recycle calculation states (`CalculationState`, matrix states).

**Consequences**:

- Drastic reduction in memory allocations
- Decreased GC pressure
- 20-30% performance improvement
- Increased code complexity

### ADR-002: Dynamic Multiplication Algorithm Selection

**Context**: FFT multiplication is more efficient than standard `math/big` for very large numbers, but has significant overhead for small numbers.

**Decision**: Implement a 2-tier `smartMultiply` function that selects the algorithm based on operand size: FFT (> 500K bits) or standard `math/big` (below). `math/big` internally uses Karatsuba for large operands.

**Consequences**:

- Optimal performance across the entire value range
- Configurable via `FFTThreshold` in `Options`
- Requires calibration for each architecture

### ADR-003: Adaptive Parallelism

**Context**: Parallelism has a synchronization cost that can exceed gains for small calculations.

**Decision**: Enable parallelism only above a configurable threshold (`ParallelThreshold`, default: 4096 bits).

**Consequences**:

- Optimal performance according to calculation size
- Avoids CPU saturation for small N
- Parallelism disabled when FFT is used (FFT already saturates CPU), re-enabled above 5M bits (`ParallelFFTThreshold`)

### ADR-004: Interface-Based Decoupling (Orchestration → CLI)

**Context**: The orchestration package was directly importing CLI packages, violating Clean Architecture principles where business logic should not depend on presentation.

**Decision**: Define `ProgressReporter` and `ResultPresenter` interfaces in the orchestration package, with implementations in the CLI package.

**Consequences**:

- Clean Architecture compliance: orchestration no longer imports CLI
- Improved testability: interfaces can be mocked for unit tests
- Flexibility: alternative presenters (JSON, TUI, GUI) can be easily added
- `NullProgressReporter` enables quiet mode without conditionals
- TUI dashboard (`internal/tui`) was added as a second implementation, validating this decoupling
- Slightly more complex initialization in the app layer

## Data Flow

```
1. config.ParseConfig() parses CLI flags + env vars → AppConfig
2. app.SetupContext() creates context with timeout + signal handling
3. fibonacci.GlobalFactory() provides calculators by name
4. orchestration.ExecuteCalculations() runs calculators concurrently via errgroup
   - Each Calculator.Calculate() delegates to FibCalculator decorator
   - FibCalculator uses CalculateWithObservers() with Freeze() for lock-free progress
   - ProgressSubject notifies ChannelObserver → progressChan
   - ProgressReporter (CLIProgressReporter or TUIProgressReporter) displays progress
5. orchestration.AnalyzeComparisonResults() compares results
6. ResultPresenter (CLIResultPresenter or TUIResultPresenter) formats and displays output
```

## Performance Considerations

1. **Zero-Allocation**: Object pools avoid allocations in critical loops
2. **Smart Parallelism**: Enabled only when beneficial
3. **2-Tier Multiplication**: FFT or standard math/big selected by operand size
4. **Strassen**: Enabled for matrices with large elements
5. **Symmetric Squaring**: Specific optimization reducing multiplications

## Extensibility

To add a new algorithm:

1. Implement the `coreCalculator` interface (`CalculateCore`, `Name`) in `internal/fibonacci/`
2. Register in `NewDefaultFactory()` in `registry.go`
3. Add corresponding tests (table-driven + golden file validation)
