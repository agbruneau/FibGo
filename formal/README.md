# Formal Verification Suite for FibGo

This directory contains formal proofs and specifications that verify the mathematical correctness and concurrency safety of the FibGo Fibonacci calculator.

## What is Verified

### Mathematical Correctness (Coq)

The Coq proofs verify that the core algorithms implemented in Go are mathematically correct:

| File | What it Proves | Go Source |
|------|---------------|-----------|
| `coq/FastDoublingCorrectness.v` | Fast Doubling identities: F(2k) = F(k)*(2*F(k+1)-F(k)), F(2k+1) = F(k+1)^2 + F(k)^2 | `internal/fibonacci/fastdoubling.go`, `internal/fibonacci/doubling_framework.go` |
| `coq/FermatArithmetic.v` | Fermat ring normalization preserves residue class mod 2^(nW)+1, carry assertion in Mul | `internal/bigfft/fermat.go` |

**FastDoublingCorrectness.v** proves:
- The standard Fibonacci recurrence F(n+2) = F(n+1) + F(n)
- The addition formula F(m+n+1) = F(m+1)*F(n+1) + F(m)*F(n)
- The Fast Doubling even identity: F(2k) = F(k) * (2*F(k+1) - F(k))
- The Fast Doubling odd identity: F(2k+1) = F(k+1)^2 + F(k)^2
- The addition step correctness: F(k+2) = F(k) + F(k+1)
- Combined iteration correctness for both bit=0 and bit=1 cases
- Termination argument for the bit-scan loop
- Subtraction safety: 2*F(k+1) >= F(k) for all k

**FermatArithmetic.v** proves:
- The fundamental property: 2^(nW) == -1 (mod 2^(nW)+1)
- norm() preserves the residue class modulo M in all three cases
- norm() maintains the representation invariant z[n] in {0, 1}
- Mul/Sqr reduction correctness: z_low + z_mid*2^(nW) + z_high*2^(2*nW) == z_low - z_mid + z_high (mod M)
- The carry assertion: after reduction, the result fits in n+1 words
- Shift, Add, and Sub operation correctness

### Concurrency Safety (TLA+)

The TLA+ specifications model the concurrent architecture and verify absence of deadlocks and race conditions:

| File | What it Models | Go Source |
|------|---------------|-----------|
| `tla/Orchestration.tla` | Concurrent calculator execution via errgroup, progress channel lifecycle | `internal/orchestration/orchestrator.go` |
| `tla/ConcurrencySemaphores.tla` | Dual semaphore interaction (task-level and FFT-level parallelism) | `internal/fibonacci/common.go`, `internal/bigfft/fft_recursion.go` |

**Orchestration.tla** verifies:
- Channel is only closed after all calculators complete (SafeClose)
- No race condition on result slots (NoRaceOnResults)
- Channel never exceeds capacity (ChannelBounded)
- Deadlock freedom and eventual termination

**ConcurrencySemaphores.tla** verifies:
- Maximum goroutines bounded by NumCPU * 3 (MaxGoroutines)
- No deadlock from independent semaphore acquisition
- Semaphore token counts are consistent with worker states
- All workers eventually complete

## Directory Structure

```
formal/
  README.md                          # This file
  coq/
    _CoqProject                      # Coq project configuration
    Makefile                         # Build automation
    FastDoublingCorrectness.v        # Fast Doubling algorithm proofs
    FermatArithmetic.v               # Fermat ring arithmetic proofs
  tla/
    README.md                        # TLA+ usage instructions
    Orchestration.tla                # Orchestrator concurrency model
    ConcurrencySemaphores.tla        # Semaphore interaction model
```

## Prerequisites

### Coq (for mathematical proofs)

- **Coq 8.18+** (tested with 8.18 and 8.19)
- Install via opam: `opam install coq`
- Or via system package manager (e.g., `apt install coq`, `brew install coq`)

### TLA+ (for concurrency specifications)

- **TLA+ Toolbox** (GUI): https://lamport.azurewebsites.net/tla/toolbox.html
- Or **tla2tools.jar** (CLI): https://github.com/tlaplus/tlaplus/releases
- Requires Java 11+

## How to Run

### Coq Proofs

```bash
cd formal/coq

# Build all proofs
make all

# Or compile individually
coqc -R . FibGo FastDoublingCorrectness.v
coqc -R . FibGo FermatArithmetic.v

# Clean build artifacts
make clean
```

A successful build produces `.vo` files with no errors, confirming all proofs type-check.

### TLA+ Model Checking

See `tla/README.md` for detailed instructions. Quick start:

```bash
# Orchestration model
java -jar tla2tools.jar -config Orchestration.cfg Orchestration.tla

# Concurrency semaphores model
java -jar tla2tools.jar -config ConcurrencySemaphores.cfg ConcurrencySemaphores.tla
```

## Relationship to Go Test Suite

The formal verification complements (but does not replace) the Go test suite:

| Aspect | Go Tests | Formal Verification |
|--------|----------|-------------------|
| **Scope** | Specific input/output pairs | Universal (all inputs) |
| **Algorithms** | Tested via golden files, fuzzing | Proven correct by construction |
| **Concurrency** | Race detector (`-race` flag) | Exhaustive state space exploration |
| **Performance** | Benchmarks, profiling | Not applicable |
| **Coverage** | Line/branch coverage metrics | Logical completeness |

The Go tests verify the implementation works correctly for tested inputs. The formal proofs verify the underlying mathematics is sound for all inputs. Together, they provide high assurance that:

1. The mathematical identities used by the algorithms are correct (Coq)
2. The Fermat ring arithmetic preserves modular equivalence (Coq)
3. The concurrent architecture is free of deadlocks and races (TLA+)
4. The implementation correctly implements these verified algorithms (Go tests)
