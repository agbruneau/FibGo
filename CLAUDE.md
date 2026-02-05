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
go generate ./...                                       # Regenerate mocks
```

Makefile targets (require `make`, not available on all systems):
```bash
make test           # go test -v -race -cover ./...
make lint           # golangci-lint run ./...
make coverage       # Generate coverage.html
make check          # format + lint + test
make security       # gosec ./...
```

## Architecture Overview

**Go Module**: `github.com/agbru/fibcalc` (Go 1.25+)

CLI-only high-performance Fibonacci calculator. Four layers:

```
Entry Point (cmd/fibcalc)
    ↓
Orchestration (internal/orchestration)  — parallel execution, result aggregation
    ↓
Business (internal/fibonacci, internal/bigfft)  — algorithms, FFT multiplication
    ↓
Presentation (internal/cli)  — progress bars, output formatting
```

### Key Interfaces and Their Relationships

**Calculator** (`internal/fibonacci/calculator.go`): Public interface consumed by orchestration. Methods: `Calculate()`, `Name()`.

**coreCalculator** (`internal/fibonacci/calculator.go`): Internal interface for algorithm implementations. Methods: `CalculateCore()`, `Name()`. Wrapped by `FibCalculator` (decorator) which adds small-N optimization and progress reporting.

**CalculatorFactory** (`internal/fibonacci/registry.go`): Creates/caches `Calculator` instances. `DefaultFactory` pre-registers "fast", "matrix", "fft". Global instance via `GlobalFactory()`. GMP calculator auto-registers via `init()` when built with `-tags=gmp`.

**MultiplicationStrategy** (`internal/fibonacci/strategy.go`): Abstraction for multiply/square operations. Strategies: `SmartStrategy` (selects Karatsuba vs FFT by operand size), `FFTStrategy` (always FFT).

**ProgressReporter / ResultPresenter** (`internal/orchestration/interfaces.go`): Decouple orchestration from presentation. CLI implementations in `internal/cli/presenter.go`. `NullProgressReporter` for quiet mode/testing.

**ProgressObserver** (`internal/fibonacci/observer.go`): Observer pattern for progress updates. `ProgressSubject` manages observers; `ChannelObserver` bridges to channel-based reporting. `CalculateWithObservers()` is the observer-aware entry point; `Calculate()` wraps it for backward compatibility.

### Core Packages

| Package | Key Files | Responsibility |
|---------|-----------|----------------|
| `internal/fibonacci` | `fastdoubling.go`, `matrix.go`, `fft_based.go` | Algorithm implementations |
| `internal/fibonacci` | `registry.go`, `calculator.go`, `strategy.go` | Factory, interfaces, strategies |
| `internal/fibonacci` | `observer.go`, `common.go`, `constants.go` | Progress, task semaphore, thresholds |
| `internal/bigfft` | `fft.go`, `fermat.go`, `pool.go`, `karatsuba.go` | FFT/Karatsuba multiplication with pooling |
| `internal/bigfft` | `arith_amd64.go`, `arith_amd64.s` | Assembly-optimized FFT (AVX2/AVX-512) |
| `internal/orchestration` | `orchestrator.go`, `interfaces.go` | Parallel execution via `errgroup` |
| `internal/cli` | `output.go`, `presenter.go`, `ui.go`, `progress_eta.go` | CLI output, progress display |
| `internal/calibration` | `calibration.go`, `adaptive.go`, `microbench.go` | Auto-tuning thresholds per hardware |
| `internal/config` | `config.go`, `env.go` | Flag parsing, env vars, validation |
| `internal/app` | `lifecycle.go`, `version.go` | Context setup (timeout + signals) |
| `internal/errors` | `errors.go`, `handler.go` | Custom error types, exit codes |
| `internal/ui` | | Color themes, `NO_COLOR` support |

### Data Flow

1. `config.ParseConfig()` parses CLI flags + env vars → `AppConfig`
2. `fibonacci.GlobalFactory()` provides calculators by name
3. `orchestration.ExecuteCalculations()` runs calculators concurrently via `errgroup`
4. Each `Calculator.Calculate()` sends `ProgressUpdate` on a channel
5. `CLIProgressReporter` displays spinner/progress bar
6. `orchestration.AnalyzeComparisonResults()` compares results
7. `CLIResultPresenter` formats and displays output

## Code Conventions

**Imports**: Group as (1) stdlib, (2) third-party, (3) internal. Error package aliased as `apperrors`.

**Error Handling**: Use `internal/errors` package (`apperrors`). Types: `ConfigError`, `CalculationError`, `TimeoutError`. Each has a standardized exit code.

**Concurrency**: Use `sync.Pool` for object recycling. Task semaphore in `common.go` limits goroutines to `runtime.NumCPU()*2`.

**Testing**: Table-driven with subtests. >75% coverage target. Golden file tests in `internal/fibonacci/testdata/fibonacci_golden.json`. Fuzz tests (`FuzzFastDoubling`). Property-based tests via `gopter`.

**Linting**: `.golangci.yml` — 20+ linters. Key limits: cyclomatic complexity 15, cognitive complexity 30, function length 100 lines / 50 statements. Relaxed in `_test.go` files.

**Commits**: [Conventional Commits](https://www.conventionalcommits.org/) — `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `chore`. Format: `<type>(<scope>): <description>`

**Branch naming**: `feature/`, `fix/`, `docs/`, `refactor/`, `perf/` prefixes.

## Key Patterns

- **Decorator**: `FibCalculator` wraps `coreCalculator` to add small-N fast path and progress reporting
- **Factory + Registry**: `DefaultFactory` with lazy creation and caching; GMP auto-registers via `init()`
- **Strategy**: `MultiplicationStrategy` selects Karatsuba vs FFT based on operand bit size
- **Observer**: `ProgressSubject`/`ProgressObserver` for progress events; `ChannelObserver` bridges to channels
- **Object Pooling**: `sync.Pool` for `big.Int` and calculation states; `MaxPooledBitLen = 4M bits` cap
- **Interface-Based Decoupling**: Orchestration depends on `ProgressReporter`/`ResultPresenter` interfaces, not CLI directly

## Build Tags & Platform-Specific Code

- **GMP**: `go build -tags=gmp` — requires libgmp. `calculator_gmp.go` auto-registers via `init()`
- **amd64 ASM**: `internal/bigfft/arith_amd64.s` — runtime CPU feature detection (AVX2/AVX-512)
- **PGO**: Profile stored at `cmd/fibcalc/default.pgo`

## Naming Conventions

**CLI Package** (`internal/cli/output.go`):
- `Display*`: Write formatted output to `io.Writer`
- `Format*`: Return formatted string, no I/O
- `Write*`: Write data to filesystem

## Adding a New Algorithm

1. Implement `coreCalculator` interface (`CalculateCore`, `Name`) in `internal/fibonacci/`
2. Register in `NewDefaultFactory()` in `registry.go`
3. Add tests (table-driven + golden file validation)

## Mock Generation

Interfaces with `//go:generate mockgen` directives:
- `Calculator` → `internal/fibonacci/mocks/mock_calculator.go`
- `MultiplicationStrategy` → `internal/fibonacci/mocks/mock_strategy.go`
- `Generator` → `internal/fibonacci/mocks/mock_generator.go`
- `Spinner` → `internal/cli/mocks/mock_ui.go`

Regenerate: `go generate ./...`

## Configuration Priority

CLI flags > Environment variables (`FIBCALC_*` prefix) > Defaults. See `.env.example`.

## Key Dependencies

| Dependency | Purpose |
|-----------|---------|
| `golang.org/x/sync` | `errgroup` for concurrent calculator execution |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/briandowns/spinner` | CLI spinner animation |
| `github.com/ncw/gmp` | GMP bindings (optional, build tag) |
| `github.com/golang/mock` | Mock generation for testing |
| `github.com/leanovate/gopter` | Property-based testing |

## Known Issues

- `Docs/TROUBLESHOOTING.md` is referenced in README but does not exist.
