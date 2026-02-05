# Fibonacci Calculator Architecture

> **Version**: 1.3.0
> **Last Updated**: January 2026

## Overview

The Fibonacci Calculator is designed according to **Clean Architecture** principles, with strict separation of responsibilities and low coupling between modules. This architecture enables maximum testability, easy scalability, and simplified maintenance.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           ENTRY POINT                                   │
│                                                                         │
│                          ┌─────────┐                                    │
│                          │   CLI   │                                    │
│                          └────┬────┘                                    │
│                               │                                         │
│                        ┌──────┴───────┐                                 │
│                        │ cmd/fibcalc  │                                 │
│                        │   main.go    │                                 │
│                        └──────┬───────┘                                 │
└───────────────────────────────┼──────────────────────────────────────────┘
                                │
┌───────────────────────────────┼──────────────────────────────────────────┐
│                   ORCHESTRATION LAYER                                    │
│                               ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    internal/orchestration                        │    │
│  │  • ExecuteCalculations() - Parallel algorithm execution         │    │
│  │  • AnalyzeComparisonResults() - Analysis and comparison         │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                            │                                            │
│  ┌─────────────────────────┼───────────────────────────────────────┐    │
│  │                         ▼                                        │    │
│  │  ┌─────────────┐  ┌─────────────┐                                │    │
│  │  │   config    │  │ calibration │                                │    │
│  │  │   Parsing   │  │   Tuning    │                                │    │
│  │  └─────────────┘  └─────────────┘                                │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└───────────────────────────────┼──────────────────────────────────────────┘
                                │
┌───────────────────────────────┼──────────────────────────────────────────┐
│                      BUSINESS LAYER                                      │
│                               ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    internal/fibonacci                            │    │
│  │                                                                  │    │
│  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────┐ │    │
│  │  │  Fast Doubling   │  │     Matrix       │  │    FFT-Based   │ │    │
│  │  │  O(log n)        │  │  Exponentiation  │  │    Doubling    │ │    │
│  │  │  Parallel        │  │  O(log n)        │  │    O(log n)    │ │    │
│  │  │  Zero-Alloc      │  │  Strassen        │  │    FFT Mul     │ │    │
│  │  └──────────────────┘  └──────────────────┘  └────────────────┘ │    │
│  │                            │                                     │    │
│  │                            ▼                                     │    │
│  │  ┌─────────────────────────────────────────────────────────────┐│    │
│  │  │                    internal/bigfft                          ││    │
│  │  │  • FFT multiplication for very large numbers                ││    │
│  │  │  • Complexity O(n log n) vs O(n^1.585) for Karatsuba        ││    │
│  │  └─────────────────────────────────────────────────────────────┘│    │
│  └─────────────────────────────────────────────────────────────────┘    │
└───────────────────────────────┼──────────────────────────────────────────┘
                                │
┌───────────────────────────────┼──────────────────────────────────────────┐
│                   PRESENTATION LAYER                                     │
│                               ▼                                         │
│  ┌──────────────────────────────────┐                                   │
│  │         internal/cli             │                                   │
│  │  • Spinner and progress bar      │                                   │
│  │  • Result formatting             │                                   │
│  │  • Colour themes                 │                                   │
│  │  • NO_COLOR support              │                                   │
│  └──────────────────────────────────┘                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

## Package Structure

### `cmd/fibcalc`

Application entry point. Responsibilities:

- Command-line argument parsing
- Component initialization
- Routing to CLI mode
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

### `internal/cli`

Command-line user interface:

- Animated spinner with progress bar
- Estimated time remaining (ETA)
- Colour theme system (dark, light, none)
- Large number formatting
- Autocompletion script generation (bash, zsh, fish, powershell)
- `NO_COLOR` environment variable support
- **Interface Implementations** (`presenter.go`):
  - `CLIProgressReporter`: Implements `orchestration.ProgressReporter` for CLI progress display
  - `CLIResultPresenter`: Implements `orchestration.ResultPresenter` for CLI result formatting

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

### ADR-003: Adaptive Parallelism

**Context**: Parallelism has a synchronization cost that can exceed gains for small calculations.

**Decision**: Enable parallelism only above a configurable threshold (`--threshold`).

**Consequences**:

- ✅ Optimal performance according to calculation size
- ✅ Avoids CPU saturation for small N
- ⚠️ Parallelism disabled when FFT is used (FFT already saturates CPU)

### ADR-004: Interface-Based Decoupling (Orchestration → CLI)

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

