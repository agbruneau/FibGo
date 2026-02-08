# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build ./cmd/fibcalc                                  # Build the CLI binary
go test -v -race -cover ./...                          # Run all tests with race detector
go test -v -short ./...                                # Skip slow tests
go test -v -run TestFastDoubling ./internal/fibonacci/  # Run single test by name
go test -bench=. -benchmem ./internal/fibonacci/        # Run benchmarks
go test -fuzz=FuzzFastDoubling ./internal/fibonacci/    # Run fuzz tests
go test -fuzz=FuzzFermatMulVsBigInt -fuzztime=10m ./internal/bigfft/  # Run fermat oracle fuzz
```

Formal verification (requires Coq 8.18+ and TLA+ Toolbox):
```bash
cd formal/coq && make                                   # Compile Coq proofs
cd formal/tla && tlc Orchestration.tla                   # Check orchestration model
cd formal/tla && tlc ConcurrencySemaphores.tla           # Check semaphore model
```

Makefile targets (require `make`, not available on all systems):
```bash
make test           # go test -v -race -cover ./...
make lint           # golangci-lint run ./...
make coverage       # Generate coverage.html
make check          # format + lint + test
make security       # gosec ./...
make benchmark      # Run benchmarks
make pgo-profile    # Generate CPU profile for PGO
make build-pgo      # Build with Profile-Guided Optimization
make build-all      # Build for Linux, Windows, macOS
```

## Architecture Overview

**Go Module**: `github.com/agbru/fibcalc` (Go 1.25+)

High-performance Fibonacci calculator with CLI and interactive TUI modes. Four layers:

```
Entry Point (cmd/fibcalc)
    ↓
Orchestration (internal/orchestration)  — parallel execution, result aggregation
    ↓
Business (internal/fibonacci, internal/bigfft)  — algorithms, FFT multiplication
    ↓
Presentation (internal/cli, internal/tui)  — CLI output or TUI dashboard
```

Support packages: `internal/calibration`, `internal/config`, `internal/app`, `internal/errors`, `internal/parallel`, `internal/format`, `internal/metrics`, `internal/ui`, `internal/sysmon`, `internal/testutil`.

### Key Interfaces and Their Relationships

**Calculator** (`internal/fibonacci/calculator.go`): Public interface consumed by orchestration. Methods: `Calculate()`, `Name()`.

**coreCalculator** (`internal/fibonacci/calculator.go`): Internal interface for algorithm implementations. Methods: `CalculateCore()`, `Name()`. Wrapped by `FibCalculator` (decorator) which adds small-N optimization (iterative for n ≤ 93) and observer-based progress reporting.

**CalculatorFactory** (`internal/fibonacci/registry.go`): Creates/caches `Calculator` instances. `DefaultFactory` pre-registers "fast", "matrix", "fft". Global instance via `GlobalFactory()`. GMP calculator auto-registers via `init()` when built with `-tags=gmp`.

**Multiplier** (`internal/fibonacci/strategy.go`): Narrow interface for multiply/square operations. Methods: `Multiply()`, `Square()`, `Name()`. Consumed by code that only needs basic arithmetic.

**DoublingStepExecutor** (`internal/fibonacci/strategy.go`): Extends `Multiplier` with `ExecuteStep()` for optimized doubling steps (e.g., FFT transform reuse). Strategies: `AdaptiveStrategy` (selects math/big vs FFT by operand size), `FFTOnlyStrategy` (always FFT), `KaratsubaStrategy` (always math/big, for testing).

**MultiplicationStrategy** (`internal/fibonacci/strategy.go`): **Deprecated** type alias for `DoublingStepExecutor`. New code should use `Multiplier` or `DoublingStepExecutor`.

**DoublingFramework** (`internal/fibonacci/doubling_framework.go`): Encapsulates the Fast Doubling loop logic (bit iteration, parallelization decisions, progress reporting). Pluggable via `DoublingStepExecutor`. Optionally supports `DynamicThresholdManager` for runtime threshold adjustment.

**MatrixFramework** (`internal/fibonacci/matrix_framework.go`): Encapsulates the Matrix Exponentiation loop (binary exponentiation, symmetric matrix squaring, Strassen switching).

**Options** (`internal/fibonacci/options.go`): Configuration struct for calculations: `ParallelThreshold`, `FFTThreshold`, `StrassenThreshold`, FFT cache settings, and dynamic threshold options. Normalized via `normalizeOptions()` to fill zero values with defaults from `constants.go`.

**DynamicThresholdManager** (`internal/fibonacci/dynamic_threshold.go`): Runtime threshold adjustment based on per-iteration timing metrics. Uses a ring buffer of `IterationMetric` records, hysteresis to prevent oscillation, and separate analysis for FFT and parallel thresholds.

**ProgressReporter / ResultPresenter** (`internal/orchestration/interfaces.go`): Decouple orchestration from presentation. CLI implementations in `internal/cli/presenter.go`. TUI implementations in `internal/tui/bridge.go`. `NullProgressReporter` for quiet mode/testing.

**ProgressObserver** (`internal/fibonacci/observer.go`): Observer pattern for progress updates. `ProgressSubject` manages observers with `Freeze()` for lock-free snapshots in hot loops. Concrete observers in `observers.go`: `ChannelObserver` (bridges to channel-based reporting), `LoggingObserver` (zerolog with throttling), `NoOpObserver` (null object).

**SequenceGenerator** (`internal/fibonacci/generator.go`): Interface for iterative/streaming Fibonacci generation. `IterativeGenerator` in `generator_iterative.go` produces consecutive terms with `Next(ctx)`.

### Core Packages

| Package | Key Files | Responsibility |
|---------|-----------|----------------|
| `cmd/fibcalc` | `main.go` | Entry point: delegates to `app.New()` and `app.Run()` |
| `cmd/generate-golden` | `main.go` | Golden file generator for test data |
| `internal/fibonacci` | `fastdoubling.go`, `matrix.go`, `fft_based.go` | Algorithm implementations (coreCalculator) |
| `internal/fibonacci` | `doubling_framework.go`, `matrix_framework.go` | Shared computation frameworks |
| `internal/fibonacci` | `registry.go`, `calculator.go`, `strategy.go` | Factory, public interfaces, strategies |
| `internal/fibonacci` | `options.go`, `constants.go` | Calculation options, default thresholds |
| `internal/fibonacci` | `observer.go`, `observers.go`, `progress.go` | Observer pattern, progress utilities |
| `internal/fibonacci` | `common.go` | Task semaphore, state pool, `executeTasks` generics |
| `internal/fibonacci` | `fft.go` | FFT wrappers (`mulFFT`, `sqrFFT`, `smartMultiply`, `smartSquare`) |
| `internal/fibonacci` | `dynamic_threshold.go`, `threshold_types.go` | Runtime threshold adjustment |
| `internal/fibonacci` | `generator.go`, `generator_iterative.go` | Sequence generation interface and implementation |
| `internal/fibonacci` | `matrix_ops.go`, `matrix_types.go` | Matrix operations and types (`matrix`, `matrixState`) |
| `internal/fibonacci` | `calculator_gmp.go` | GMP calculator (build tag `gmp`) |
| `internal/fibonacci` | `testing.go` | Test helpers (exported for test packages) |
| `internal/bigfft` | `fft.go`, `fft_core.go`, `fft_recursion.go` | FFT multiplication core |
| `internal/bigfft` | `fft_poly.go` | Polynomial operations for FFT |
| `internal/bigfft` | `fft_cache.go` | Thread-safe LRU cache for FFT transforms |
| `internal/bigfft` | `fermat.go` | Fermat number arithmetic |
| `internal/bigfft` | `pool.go`, `pool_warming.go` | `sync.Pool` for `big.Int` and pre-warming |
| `internal/bigfft` | `bump.go` | O(1) bump allocator for FFT temporaries |
| `internal/bigfft` | `allocator.go` | `TempAllocator` interface (`PoolAllocator`, `BumpAllocator`) |
| `internal/bigfft` | `memory_est.go`, `scan.go` | Memory estimation, scanning utilities |
| `internal/bigfft` | `arith_amd64.go`, `arith_generic.go`, `arith_decl.go` | Vector arithmetic via `go:linkname` to `math/big` |
| `internal/bigfft` | `cpu_amd64.go` | Runtime CPU feature detection (AVX2/AVX-512) |
| `internal/orchestration` | `orchestrator.go`, `interfaces.go` | Parallel execution via `errgroup`, result analysis |
| `internal/orchestration` | `calculator_selection.go` | Calculator selection logic from config |
| `internal/cli` | `output.go`, `presenter.go` | CLI output formatting, result presentation |
| `internal/cli` | `ui.go`, `ui_display.go`, `ui_format.go` | UI helpers (spinners, formatting, display) |
| `internal/cli` | `progress_eta.go` | ETA calculation and progress display |
| `internal/cli` | `completion.go` | Shell completion generation (bash, zsh, fish, powershell) |
| `internal/cli` | `provider.go` | Progress reporter and config display providers |
| `internal/cli` | `calculate.go` | Calculation result display helpers |
| `internal/tui` | `model.go` | Bubble Tea model (Elm architecture), main Update/View |
| `internal/tui` | `bridge.go` | `TUIProgressReporter` / `TUIResultPresenter` adapters |
| `internal/tui` | `header.go`, `footer.go` | Header (title, elapsed) and footer (keys, status) panels |
| `internal/tui` | `logs.go` | Scrollable calculation log panel |
| `internal/tui` | `metrics.go` | Runtime metrics panel (memory, GC, goroutines) |
| `internal/tui` | `chart.go`, `sparkline.go` | Progress chart and sparkline visualization |
| `internal/tui` | `styles.go`, `keymap.go`, `messages.go` | Lipgloss styles, key bindings, Bubble Tea messages |
| `internal/calibration` | `calibration.go`, `runner.go` | Full calibration mode, benchmark runner |
| `internal/calibration` | `adaptive.go` | Hardware-adaptive threshold estimation |
| `internal/calibration` | `profile.go`, `io.go` | Calibration profile persistence (JSON) |
| `internal/calibration` | `microbench.go` | Micro-benchmarks for threshold determination |
| `internal/config` | `config.go`, `env.go`, `usage.go` | Flag parsing, env var overrides, custom usage |
| `internal/app` | `app.go`, `version.go` | Application lifecycle, dispatching, version info |
| `internal/errors` | `errors.go`, `handler.go` | Custom error types (`ConfigError`, `CalculationError`), exit codes |
| `internal/parallel` | `errors.go` | `ErrorCollector` for concurrent error aggregation |
| `internal/format` | `duration.go`, `numbers.go`, `progress_eta.go` | Duration/number formatting, ETA display (shared by CLI and TUI) |
| `internal/metrics` | `indicators.go` | Performance indicators (bits/s, digits/s, steps/s) |
| `internal/sysmon` | `sysmon.go` | System-wide CPU/memory monitoring via gopsutil |
| `internal/ui` | `colors.go`, `themes.go` | Color themes, `NO_COLOR` support |
| `internal/testutil` | `ansi.go` | ANSI escape code stripping for test assertions |
| `formal/coq` | `FastDoublingCorrectness.v`, `FermatArithmetic.v` | Machine-checked Coq proofs of algorithm identities and Fermat arithmetic |
| `formal/tla` | `Orchestration.tla`, `ConcurrencySemaphores.tla` | TLA+ specifications for concurrency model verification |

### Data Flow

1. `app.New()` calls `config.ParseConfig()` to parse CLI flags + env vars → `AppConfig`
2. `calibration.LoadCachedCalibration()` attempts to load stored calibration profile
3. If no profile, `applyAdaptiveThresholds()` estimates thresholds from hardware characteristics
4. `app.Run()` dispatches to completion, calibration, TUI, or CLI mode
5. `orchestration.GetCalculatorsToRun()` selects calculators from `fibonacci.GlobalFactory()`
6. `orchestration.ExecuteCalculations()` runs calculators concurrently via `errgroup`
7. Each `Calculator.Calculate()` creates a `ProgressSubject`, registers a `ChannelObserver`, and delegates to `CalculateWithObservers()`
8. `DoublingFramework.ExecuteDoublingLoop()` or `MatrixFramework.ExecuteMatrixLoop()` runs the core algorithm
9. `CLIProgressReporter` or `TUIProgressReporter` displays progress from the channel
10. `orchestration.AnalyzeComparisonResults()` compares results and sorts by speed
11. `CLIResultPresenter` or `TUIResultPresenter` formats and displays output

## Code Conventions

**Imports**: Group as (1) stdlib, (2) third-party, (3) internal. Error package aliased as `apperrors`.

**Error Handling**: Use `internal/errors` package (`apperrors`). Types: `ConfigError`, `CalculationError`. Exit codes: `ExitSuccess` (0), `ExitErrorGeneric` (1), `ExitErrorTimeout` (2), `ExitErrorMismatch` (3), `ExitErrorConfig` (4), `ExitErrorCanceled` (130). Helper: `WrapError()`, `IsContextError()`.

**Concurrency**: Use `sync.Pool` for object recycling. Task semaphore in `common.go` limits goroutines to `runtime.NumCPU()*2`. `parallel.ErrorCollector` for first-error aggregation.

**Testing**: Table-driven with subtests. >75% coverage target. Golden file tests in `internal/fibonacci/testdata/fibonacci_golden.json`. Fuzz tests (`FuzzFastDoubling`). Property-based tests via `gopter`. Example tests in `example_test.go`. E2E tests in `test/e2e/`. Formal verification: 17 oracle-based fuzz targets (`fermat_fuzz_test.go`, `fft_roundtrip_fuzz_test.go`, `fibonacci_formal_fuzz_test.go`, `strategy_oracle_test.go`), exhaustive state aliasing tests (`state_aliasing_test.go`), concurrency verification tests (`orchestration_deadlock_test.go`, `observer_concurrency_test.go`, `semaphore_verification_test.go`, `errors_concurrency_test.go`).

**Linting**: `.golangci.yml` — 22 linters enabled. Key limits: cyclomatic complexity 15, cognitive complexity 30, function length 100 lines / 50 statements. Relaxed in `_test.go` files.

**Commits**: [Conventional Commits](https://www.conventionalcommits.org/) — `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `chore`. Format: `<type>(<scope>): <description>`

**Branch naming**: `feature/`, `fix/`, `docs/`, `refactor/`, `perf/` prefixes.

## Key Patterns

- **Decorator**: `FibCalculator` wraps `coreCalculator` to add small-N fast path (iterative for n ≤ 93) and observer-based progress reporting
- **Factory + Registry**: `DefaultFactory` with lazy creation and caching; GMP auto-registers via `init()`
- **Strategy + ISP**: `Multiplier` (narrow) and `DoublingStepExecutor` (wide) interfaces; `AdaptiveStrategy`, `FFTOnlyStrategy`, `KaratsubaStrategy` implementations
- **Framework**: `DoublingFramework` and `MatrixFramework` encapsulate algorithm loops, decoupled from multiplication strategies
- **Observer**: `ProgressSubject`/`ProgressObserver` for progress events; `Freeze()` creates lock-free snapshots for hot loops; concrete: `ChannelObserver`, `LoggingObserver`, `NoOpObserver`
- **Object Pooling**: `sync.Pool` for `big.Int`, `CalculationState`, and `matrixState`; `MaxPooledBitLen = 100M bits` cap to prevent oversized objects staying in pool
- **Bump Allocator**: O(1) temporary allocation for FFT operations via pointer bump; zero fragmentation and excellent cache locality
- **FFT Transform Cache**: Thread-safe LRU cache in `bigfft/fft_cache.go`; configurable via `TransformCacheConfig`; 15-30% speedup for iterative algorithms
- **Dynamic Threshold Adjustment**: `DynamicThresholdManager` monitors per-iteration timing, adjusts FFT/parallel thresholds with hysteresis to prevent oscillation
- **Zero-Copy Result Return**: Algorithms "steal" the result pointer from pooled state, replacing it with a fresh `big.Int`, avoiding O(n) copy
- **Interface-Based Decoupling**: Orchestration depends on `ProgressReporter`/`ResultPresenter` interfaces, not CLI directly
- **Generics**: `executeTasks[T, PT]()` in `common.go` uses Go generics with pointer constraint pattern to eliminate duplication between multiplication and squaring tasks

## Build Tags & Platform-Specific Code

- **GMP**: `go build -tags=gmp` — requires libgmp. `calculator_gmp.go` auto-registers via `init()`
- **amd64 optimizations**: `internal/bigfft/arith_amd64.go` — delegates to `math/big` assembly via `go:linkname`; `cpu_amd64.go` — runtime CPU feature detection (AVX2/AVX-512)
- **Generic fallback**: `internal/bigfft/arith_generic.go` — portable implementations for non-amd64
- **PGO**: Profile stored at `cmd/fibcalc/default.pgo`; auto-used by `make build` if present

## Naming Conventions

**CLI Package** (`internal/cli/output.go`):
- `Display*`: Write formatted output to `io.Writer`
- `Format*`: Return formatted string, no I/O
- `Write*`: Write data to filesystem
- `Print*`: Write to stdout (convenience wrappers)

## Adding a New Algorithm

1. Implement `coreCalculator` interface (`CalculateCore`, `Name`) in `internal/fibonacci/`
2. Register in `NewDefaultFactory()` in `registry.go`
3. Add tests (table-driven + golden file validation)

## Configuration Priority

CLI flags > Environment variables (`FIBCALC_*` prefix) > Adaptive hardware estimation > Static defaults. See `.env.example`.

**Threshold resolution**: CLI flag defaults are `0` (auto). When `0`, the app first tries to load a cached calibration profile, then falls back to `calibration.EstimateOptimal*Threshold()` functions that adapt based on CPU core count and architecture. Static defaults in `constants.go` (`DefaultParallelThreshold=4096`, `DefaultFFTThreshold=500,000`, `DefaultStrassenThreshold=3072`) are used inside algorithm code when `normalizeOptions()` encounters zero values.

## Key Dependencies

| Dependency | Purpose |
|-----------|---------|
| `golang.org/x/sync` | `errgroup` for concurrent calculator execution |
| `golang.org/x/sys` | System calls (signal handling, platform detection) |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/briandowns/spinner` | CLI spinner animation |
| `github.com/shirou/gopsutil/v4` | System metrics (CPU/memory usage for TUI) |
| `github.com/ncw/gmp` | GMP bindings (optional, build tag `gmp`) |
| `go.uber.org/mock` | Mock generation for testing |
| `github.com/leanovate/gopter` | Property-based testing |
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `github.com/charmbracelet/lipgloss` | TUI styling and layout |
| `github.com/charmbracelet/bubbles` | TUI components (key bindings, viewport) |
