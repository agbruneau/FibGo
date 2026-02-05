# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **CLI-only refactor**: Removed server, TUI, REPL, and observability layers to simplify the codebase
- Cleaned up documentation to reflect CLI-only architecture

---

## [1.0.0] - 2025-12-22

### Added

#### Core Features

- **Fast Doubling Algorithm**: O(log n) Fibonacci calculation with parallel multiplication
- **Matrix Exponentiation**: O(log n) with Strassen's algorithm for large matrices
- **FFT-Based Calculator**: Optimized for extremely large numbers using FFT multiplication
- **GMP Support**: Optional GNU Multiple Precision library integration via build tag

#### Performance Optimizations

- Zero-allocation strategy using `sync.Pool` for 95%+ reduction in GC pressure
- Adaptive parallelism based on input size and hardware capabilities
- Smart multiplication switching (Karatsuba vs FFT) based on operand size
- Symmetric matrix squaring optimization (50% reduction in multiplications)
- Auto-calibration system for hardware-specific threshold optimization

#### User Interface

- Modern CLI with progress spinners, ETA calculation, and colour themes
- Shell autocompletion generation (bash, zsh, fish, PowerShell)
- JSON output format support
- Hexadecimal result display option

#### Documentation

- Comprehensive README with production deployment guide
- Architecture documentation with ADRs
- Performance tuning guide
- Security policy with vulnerability disclosure process
- Algorithm-specific documentation (Fast Doubling, Matrix, FFT, GMP)

#### Development

- Comprehensive test suite with 80%+ coverage
- Benchmark suite for performance validation
- Mock generation with mockgen
- golangci-lint configuration

### Security

- Input validation for all parameters
- Maximum N value limit (1 billion) to prevent resource exhaustion
- Configurable request timeouts
- Rate limiting protection against DoS

---

## [0.1.0] - 2025-11-01

### Added

- Initial project structure
- Basic Fast Doubling implementation
- Command-line interface
- Unit tests for core algorithms

---

[Unreleased]: https://github.com/agbru/fibcalc/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/agbru/fibcalc/compare/v0.1.0...v1.0.0
[0.1.0]: https://github.com/agbru/fibcalc/releases/tag/v0.1.0
