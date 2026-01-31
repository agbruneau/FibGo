# FibCalc: High-Performance Fibonacci Calculator

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=for-the-badge&logo=apache)

**FibCalc** is a command-line tool for computing arbitrarily large Fibonacci numbers with extreme speed and efficiency. Written in Go, it leverages advanced algorithmic optimizations -- including Fast Doubling, Matrix Exponentiation with Strassen's algorithm, and FFT-based multiplication -- to handle indices in the hundreds of millions.

---

## Key Features

### Algorithms

- **Fast Doubling** (Default): $O(\log n)$ using $F(2k) = F(k)(2F(k+1) - F(k))$.
- **Matrix Exponentiation**: $O(\log n)$ with Strassen's algorithm for large matrices.
- **FFT-Based Multiplication**: Switches to FFT when numbers exceed ~500k bits, reducing complexity from $O(n^{1.585})$ to $O(n \log n)$.
- **GMP Support**: Optional build tag for GNU Multiple Precision Arithmetic Library.

### Performance

- **Object Pooling**: `sync.Pool` for `big.Int` recycling, reducing GC pressure by over 95%.
- **Adaptive Parallelism**: Automatic parallelization based on input size and hardware.
- **Concurrency Limiting**: Task semaphore capped at `runtime.NumCPU()*2`.

---

## Quick Start

```bash
# Build
make build

# Calculate the 10-millionth Fibonacci number
./build/fibcalc -n 10000000

# Interactive REPL mode
./build/fibcalc --interactive
```

---

## Installation

Requires **Go 1.25** or later.

```bash
git clone https://github.com/agbru/fibcalc.git
cd fibcalc
make build
# Binary is located at ./build/fibcalc
```

---

## Usage

```text
fibcalc [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--n` | `-n` | `250,000,000` | Fibonacci index to calculate. |
| `--algo` | | `all` | Algorithm: `fast`, `matrix`, `fft`, or `all`. |
| `--output` | `-o` | | Write result to a file. |
| `--hex` | | `false` | Display result in hexadecimal. |
| `--calculate` | `-c` | `false` | Print the full value. |
| `--interactive` | | `false` | Start interactive REPL mode. |
| `--timeout` | | `5m` | Maximum calculation time. |
| `--threshold` | | `4096` | Parallelism threshold in bits. |
| `--fft-threshold` | | `500000` | FFT multiplication threshold in bits. |
| `--strassen-threshold` | | `3072` | Strassen algorithm threshold in bits. |
| `--quiet` | `-q` | `false` | Quiet mode for scripting. |
| `--no-color` | | `false` | Disable colored output. |
| `--completion` | | | Generate shell completion script (bash, zsh, fish, powershell). |

### Examples

```bash
# Compare all algorithms with detailed stats
fibcalc -n 10000000 --algo all --details

# Interactive session
fibcalc --interactive
# fib> calc 100
# fib> algo matrix
# fib> compare 50000
# fib> exit

# Force FFT with lower threshold
fibcalc -n 5000000 --algo fast --fft-threshold 100000

# Quiet mode for scripting
fibcalc -n 1000 --quiet
```

---

## Architecture

```
internal/
├── fibonacci/   # Core algorithms (Fast Doubling, Matrix, FFT-based)
├── bigfft/      # FFT multiplication for big.Int
├── cli/         # CLI output, REPL, progress bar, spinner
├── errors/      # Custom error types with exit codes
├── parallel/    # Concurrency utilities
└── ui/          # Terminal colors, NO_COLOR support
```

| Package | Responsibility |
|---------|----------------|
| `internal/fibonacci` | Calculator interface, algorithm implementations, strategy pattern |
| `internal/bigfft` | FFT arithmetic for `big.Int` with memory pooling |
| `internal/cli` | REPL, progress bar, spinner, output formatting |
| `internal/errors` | Structured error types with standardized exit codes |
| `internal/parallel` | Concurrency utilities |
| `internal/ui` | ANSI color codes, `NO_COLOR` support |

---

## Performance Benchmarks

Results on AMD Ryzen 9 5900X:

| Index ($N$) | Fast Doubling | Matrix Exp. | FFT-Based | Digits |
| :--- | :--- | :--- | :--- | :--- |
| **10,000** | 180us | 220us | 350us | 2,090 |
| **1,000,000** | 85ms | 110ms | 95ms | 208,988 |
| **100,000,000** | 45s | 62s | 48s | 20,898,764 |
| **250,000,000** | 3m 12s | 4m 25s | 3m 28s | 52,246,909 |

### Algorithm Selection

- **`fast`** (Fast Doubling): Best general-purpose performance across all ranges.
- **`matrix`**: Educational purposes or verification.
- **`fft`**: Competitive for $N > 100,000,000$.

---

## Development

### Prerequisites
- Go 1.25+
- Make

### Commands

```bash
make build       # Compile binary to ./build/fibcalc
make test        # Run all tests with race detector
make test-short  # Run tests without slow ones
make lint        # Run golangci-lint
make check       # Run format + lint + test
make coverage    # Generate coverage report (coverage.html)
make benchmark   # Run performance benchmarks
make clean       # Remove build artifacts
```

---

## Troubleshooting

### `runtime: out of memory`
Large Fibonacci numbers require significant RAM. $F(1,000,000,000)$ needs ~25 GB.
Reduce $N$, add swap space, or use a machine with more RAM.

### Calculation timeout
For very large $N$, increase the timeout: `--timeout 30m`.

---

## License

This project is licensed under the Apache License 2.0 -- see the [LICENSE](LICENSE) file for details.
