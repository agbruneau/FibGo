# TLA+ Specifications for FibGo

This directory contains TLA+ specifications that model the concurrency architecture of the FibGo Fibonacci calculator.

## Specifications

### Orchestration.tla

Models the concurrent execution lifecycle from `internal/orchestration/orchestrator.go`:

- **N calculators** run concurrently via `errgroup`
- A **bounded progress channel** carries progress updates (capacity = N * 50)
- A **display goroutine** consumes updates from the channel
- The channel is **closed after all calculators complete** (`g.Wait()` then `close()`)
- The display goroutine **drains remaining messages** then terminates

**Properties verified:**

| Property | Type | Description |
|----------|------|-------------|
| `SafeClose` | Safety | Channel is only closed after all calculators are done |
| `NoRaceOnResults` | Safety | Each result slot is written exactly once |
| `ChannelBounded` | Safety | Channel never exceeds its capacity |
| `ResultsOnlyFromDone` | Safety | Results are only written by completed calculators |
| `DeadlockFreedom` | Liveness | All calculators finish and display terminates |
| `Termination` | Liveness | The system eventually reaches a terminal state |
| `CalculatorProgress` | Liveness | Every started calculator eventually completes |
| `ChannelEventuallyDrained` | Liveness | Closed channel is eventually fully drained |

### ConcurrencySemaphores.tla

Models the dual semaphore interaction between the Fibonacci task-level parallelism (`internal/fibonacci/common.go`) and the FFT recursion-level parallelism (`internal/bigfft/fft_recursion.go`):

- **Task Semaphore**: capacity = NumCPU * 2, limits Fibonacci-level goroutines
- **FFT Semaphore**: capacity = NumCPU, limits FFT recursive goroutines
- Workers acquire task semaphore first, then optionally acquire FFT semaphore
- Total ordering on acquisition prevents deadlock

**Properties verified:**

| Property | Type | Description |
|----------|------|-------------|
| `SemaphoreBounds` | Safety | Semaphore counts stay within capacity |
| `MaxGoroutines` | Safety | Total active goroutines <= NumCPU * 3 |
| `TaskSemConsistency` | Safety | Token count matches worker states |
| `FFTSemConsistency` | Safety | Token count matches worker states |
| `NoDeadlock` | Safety | No state where all workers are blocked |
| `EventualCompletion` | Liveness | Every worker eventually completes |
| `AllDone` | Liveness | All workers eventually finish |
| `SemaphoresReleased` | Liveness | Both semaphores are eventually fully released |

## Prerequisites

Install one of:

1. **TLA+ Toolbox** (GUI): Download from https://lamport.azurewebsites.net/tla/toolbox.html
2. **TLC command-line** (part of the tla2tools.jar): Download from https://github.com/tlaplus/tlaplus/releases

## Running with TLA+ Toolbox

1. Open the TLA+ Toolbox
2. Create a new specification and add the `.tla` file
3. Create a new model:
   - For `Orchestration.tla`:
     - Set `NumCalculators = 3` (or 2 for faster checking)
     - Set `ChannelCapacity = 6` (NumCalculators * 2, reduced for model checking)
   - For `ConcurrencySemaphores.tla`:
     - Set `NumCPU = 2`
     - Set `NumWorkers = 4`
     - Set `TaskSemCapacity = 4` (NumCPU * 2)
     - Set `FFTSemCapacity = 2` (NumCPU)
4. Add the invariants and temporal properties to check
5. Run TLC

## Running with Command-Line TLC

Create a configuration file (e.g., `Orchestration.cfg`):

```
CONSTANTS
    NumCalculators = 3
    ChannelCapacity = 6

SPECIFICATION Spec

INVARIANTS
    SafetyInvariant

PROPERTIES
    DeadlockFreedom
    Termination
```

Then run:

```bash
java -jar tla2tools.jar -config Orchestration.cfg Orchestration.tla
```

For `ConcurrencySemaphores.tla`, create `ConcurrencySemaphores.cfg`:

```
CONSTANTS
    NumCPU = 2
    NumWorkers = 4
    TaskSemCapacity = 4
    FFTSemCapacity = 2

SPECIFICATION Spec

INVARIANTS
    SafetyInvariant

PROPERTIES
    AllDone
    SemaphoresReleased
```

Then run:

```bash
java -jar tla2tools.jar -config ConcurrencySemaphores.cfg ConcurrencySemaphores.tla
```

## Expected Output

When all properties are satisfied, TLC will output:

```
Model checking completed. No error has been found.
  Finished in XXs at (date)
  NNNNN states generated, NNNN distinct states found, 0 states left on queue.
```

If a property is violated, TLC will produce a counterexample trace showing the sequence of states that leads to the violation.

## Notes on State Space

- Keep constants small for model checking (e.g., NumCalculators <= 3, NumWorkers <= 4)
- Larger values exponentially increase the state space
- The small models are sufficient to find concurrency bugs because TLA+ exhaustively explores all interleavings
- The properties proven for small N generalize to arbitrary N by the symmetry of the specification
