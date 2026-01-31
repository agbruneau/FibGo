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
```

## Architecture Overview

**Go Module**: `github.com/agbru/fibcalc` (Go 1.25+)

This is a high-performance Fibonacci calculator implementing multiple algorithms with Clean Architecture principles. The codebase has four main layers:

### Entry Points → Orchestration → Business → Presentation

1. **Entry Points** (`cmd/fibcalc`): CLI main, routes to CLI/Server/REPL/TUI modes
2. **Orchestration** (`internal/orchestration`): Parallel algorithm execution, result aggregation
3. **Business** (`internal/fibonacci`, `internal/bigfft`): Core algorithms and FFT multiplication
4. **Presentation** (`internal/cli`, `internal/tui`, `internal/server`): User interface, TUI, and HTTP API

### Core Packages

| Package | Responsibility |
|---------|----------------|
| `internal/fibonacci` | Calculator interface, algorithms (Fast Doubling, Matrix, FFT-based) |
| `internal/bigfft` | FFT multiplication for large `big.Int` - O(n log n) vs Karatsuba O(n^1.585) |
| `internal/orchestration` | Concurrent algorithm execution with timeout/cancellation |
| `internal/server` | REST API: `/calculate`, `/health`, `/algorithms`, `/metrics` |
| `internal/cli` | REPL, spinner, progress bar with ETA, color themes |
| `internal/tui` | Rich Terminal UI using Bubbletea (navigation, progress, theming) |
| `internal/calibration` | Auto-tuning to find optimal thresholds per hardware |
| `internal/config` | Configuration management and validation |
| `internal/service` | Business logic layer |
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

**Linting**: `.golangci.yml` enforces gofmt, govet, errcheck, staticcheck, revive, gosec (cyclomatic complexity max 15)

## Key Patterns

- **Strategy Pattern**: Calculator interface abstracts algorithm implementations
- **Object Pooling**: `sync.Pool` for `big.Int` and calculation states (20-30% perf gain)
- **Smart Multiplication**: `smartMultiply` selects Karatsuba vs FFT based on operand size
- **Adaptive Parallelism**: Parallelism enabled only above configurable threshold
- **Task Semaphore**: Limits concurrent goroutines to `runtime.NumCPU()*2` in `internal/fibonacci/common.go`
- **Optimized Zeroing**: Uses Go 1.21+ `clear()` builtin instead of loops in `internal/bigfft`
- **Interface-Based Decoupling**: Orchestration layer uses `ProgressReporter` and `ResultPresenter` interfaces (defined in `internal/orchestration/interfaces.go`) to avoid depending on CLI. Implementations in `internal/cli/presenter.go` and `internal/tui/presenter.go`
- **Elm Architecture (TUI)**: The TUI follows Bubbletea's Elm-inspired Model-Update-View pattern with message-based state updates

## Naming Conventions

**CLI Package (`internal/cli/output.go`)**:
- `Display*` functions: Write formatted output to `io.Writer` (e.g., `DisplayResult`)
- `Format*` functions: Return formatted string, no I/O (e.g., `FormatQuietResult`)
- `Write*` functions: Write data to filesystem (e.g., `WriteResultToFile`)

**TUI Package (`internal/tui/`)** - HTOP-style single-screen dashboard:
- `dashboard.go`: Main DashboardModel with Init, Update, View
- `dashboard_input.go`: Input section (N field, calculate/compare buttons)
- `dashboard_algorithms.go`: Algorithm table with real-time progress bars
- `dashboard_results.go`: Results display section
- `dashboard_overlays.go`: Help overlay and settings panel
- `sections.go`: Section type (Input, Algorithms, Results) and navigation
- `*Msg` types: Messages for state updates (ProgressMsg, ResultMsg, etc.)
- `*State` structs: Consolidated state (InputState, AlgorithmTableState, etc.)

## Adding New Components

**New Algorithm**: Implement `coreCalculator` interface in `internal/fibonacci`, register in `calculatorRegistry`

**New API Endpoint**: Add handler in `internal/server/server.go`, register route in `NewServer()`, update OpenAPI docs

## Test Coverage

Key test files added during refactorisation:
- `internal/logging/logger_test.go` - Tests for logging adapters and field helpers
- `internal/server/security_test.go` - Tests for CORS headers and SecurityMiddleware
- `internal/server/metrics_test.go` - Tests for Prometheus metrics

Run specific tests: `go test -v -run <TestName> ./internal/<package>/`

## Key Dependencies

prometheus/client_golang, zerolog, go.opentelemetry.io/otel, golang.org/x/sync, gmp

**TUI Dependencies** (Charm stack):
- `github.com/charmbracelet/bubbletea` - Elm-inspired TUI framework
- `github.com/charmbracelet/bubbles` - Common UI components (help, key bindings)
- `github.com/charmbracelet/lipgloss` - Style definitions and rendering
