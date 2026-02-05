# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build              # Build binary to ./build/fibcalc
make test               # Run all tests with race detector
make test-short         # Run tests without slow ones
go test -v -run <TEST> ./internal/fibonacci/  # Run single test by name
make coverage           # Generate coverage report (coverage.html)
make benchmark          # Run benchmarks for fibonacci algorithms
make lint               # Run golangci-lint
make check              # Run format + lint + test
make clean              # Remove build artifacts
make generate-mocks     # Regenerate mock implementations
make security           # Run gosec security audit
make pgo-rebuild        # Full PGO workflow: profile + build
go test -fuzz=FuzzFastDoubling ./internal/fibonacci/  # Run fuzz tests
```

## Architecture Overview

**Go Module**: `github.com/agbru/fibcalc` (Go 1.25+)

This is a high-performance Fibonacci calculator implementing multiple algorithms with Clean Architecture principles. The codebase has four main layers:

### Entry Points → Orchestration → Business → Presentation

1. **Entry Points** (`cmd/fibcalc`): CLI main, routes to CLI mode
2. **Orchestration** (`internal/orchestration`): Parallel algorithm execution, result aggregation
3. **Business** (`internal/fibonacci`, `internal/bigfft`): Core algorithms and FFT multiplication
4. **Presentation** (`internal/cli`): CLI output, progress bars, shell completions

### Core Packages

| Package | Responsibility |
|---------|----------------|
| `internal/fibonacci` | Calculator interface, algorithms (Fast Doubling, Matrix, FFT-based) |
| `internal/bigfft` | FFT multiplication for large `big.Int` - O(n log n) vs Karatsuba O(n^1.585) |
| `internal/orchestration` | Concurrent algorithm execution with timeout/cancellation |
| `internal/cli` | Spinner, progress bar with ETA, color themes, shell completions |
| `internal/calibration` | Auto-tuning to find optimal thresholds per hardware |
| `internal/config` | Configuration management and validation |
| `internal/parallel` | Concurrency utilities |
| `internal/errors` | Custom error types with standardized exit codes |
| `internal/app` | Application composition root and lifecycle management |
| `internal/ui` | Color themes, terminal formatting, NO_COLOR support |
| `internal/logging` | Structured logging with zerolog adapters |

### Key Algorithms

- **Fast Doubling** (default): O(log n) using F(2k) = F(k)(2F(k+1) - F(k))
- **Matrix Exponentiation**: O(log n) with Strassen's algorithm for large matrices
- **FFT-Based**: Switches to FFT multiplication when numbers exceed ~500k bits

## Code Conventions

**Imports**: Group as (1) stdlib, (2) third-party, (3) internal

**Error Handling**: Use `internal/errors` package; always wrap errors

**Concurrency**: Use `sync.Pool` for object recycling to minimize GC pressure

**Testing**: Table-driven tests with subtests; >75% coverage target; use mockgen for mocks

**Configuration**: Use functional options pattern for configurable components

**Linting**: `.golangci.yml` enforces gofmt, govet, errcheck, staticcheck, revive, gosec, and 20+ more linters. Key thresholds: cyclomatic complexity max 15, cognitive complexity max 30, function length max 100 lines / 50 statements. Complexity/length linters are relaxed in test files.

**Commits**: Follow [Conventional Commits](https://www.conventionalcommits.org/) — `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`. Format: `<type>(<scope>): <description>`

**Branch naming**: `feature/`, `fix/`, `docs/`, `refactor/`, `perf/` prefixes (e.g., `feature/add-new-algorithm`)

## Key Patterns

- **Strategy Pattern**: Calculator interface abstracts algorithm implementations
- **Object Pooling**: `sync.Pool` for `big.Int` and calculation states (20-30% perf gain)
- **Smart Multiplication**: `smartMultiply` selects Karatsuba vs FFT based on operand size
- **Adaptive Parallelism**: Parallelism enabled only above configurable threshold
- **Task Semaphore**: Limits concurrent goroutines to `runtime.NumCPU()*2` in `internal/fibonacci/common.go`
- **Optimized Zeroing**: Uses Go 1.21+ `clear()` builtin instead of loops in `internal/bigfft`
- **Interface-Based Decoupling**: Orchestration layer uses `ProgressReporter` and `ResultPresenter` interfaces (defined in `internal/orchestration/interfaces.go`) to avoid depending on CLI. Implementation in `internal/cli/presenter.go`
- **Observer Pattern**: `ProgressObserver` interface (`internal/fibonacci/observer.go`) enables multiple progress consumers (UI, logging). Concrete observers: `ChannelObserver`, `LoggingObserver`. Calculators expose `CalculateWithObservers()` alongside standard `Calculate()`

## Build Tags & Platform-Specific Code

- **GMP support**: Build with `-tags=gmp` to use GNU Multiple Precision Arithmetic Library via `internal/fibonacci/calculator_gmp.go`. Auto-registers via `init()`.
- **amd64 optimizations**: `internal/bigfft/arith_amd64.go` and `arith_amd64.s` provide assembly-optimized FFT operations with runtime CPU feature detection (AVX2/AVX-512 dynamic dispatch)
- **PGO**: Profile-Guided Optimization via `make pgo-profile` then `make build-pgo`. Profile stored at `cmd/fibcalc/default.pgo`

## Naming Conventions

**CLI Package (`internal/cli/output.go`)**:
- `Display*` functions: Write formatted output to `io.Writer` (e.g., `DisplayResult`)
- `Format*` functions: Return formatted string, no I/O (e.g., `FormatQuietResult`)
- `Write*` functions: Write data to filesystem (e.g., `WriteResultToFile`)

## Adding New Components

**New Algorithm**: Implement `coreCalculator` interface in `internal/fibonacci`, register in `calculatorRegistry`

**Mock Locations**: Mocks live in `mocks/` subdirectories (e.g., `internal/fibonacci/mocks/mock_calculator.go`). Regenerate with `make generate-mocks` after modifying interfaces.

## Configuration Priority

CLI flags > Environment variables > Defaults. See `.env.example` for all `FIBCALC_*` environment variables.

## Key Dependencies

zerolog, golang.org/x/sync, gmp
