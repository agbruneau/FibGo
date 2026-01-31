# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Rich Terminal User Interface (TUI)

- **TUI Mode** (`--tui`): Full-featured Terminal User Interface built with the Charm stack (Bubbletea, Bubbles, Lipgloss)
- **Multiple Views**: Home, Calculator, Progress, Results, Comparison, Settings, Help screens
- **Real-time Progress**: Animated progress bar with ETA calculation during calculations
- **Algorithm Comparison**: Side-by-side multi-algorithm benchmark view
- **Theme Support**: Dark, light, and no-color themes (respects `NO_COLOR` environment variable)
- **Keyboard Navigation**: Full keyboard shortcuts for all operations
- **Result Actions**: Save to file, toggle hexadecimal display, view full result
- **Settings Screen**: Configure default algorithm, theme, and display options
- **Elm Architecture**: Clean Model-Update-View pattern for maintainability

#### Documentation

- Documentation gap analysis and improvements for production readiness
- TUI section in README.md with full feature documentation
- TUI architecture documentation in ARCHITECTURE.md
- TUI troubleshooting section in TROUBLESHOOTING.md
- TUI development guidelines in CONTRIBUTING.md

### Fixed

#### TUI Display Improvements

- **Button Navigation**: Added left/right arrow key navigation between Calculate and Compare buttons in input section
- **Button Focus Indicator**: Both Calculate and Compare buttons now correctly display focus state
- **Adaptive Table Separator**: Algorithm table separator now adapts to terminal width instead of fixed 100 characters
- **Responsive Header**: Header gracefully degrades on narrow terminals (shows "[?] Help" when space is limited)

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

- Interactive REPL mode with commands: `calc`, `algo`, `compare`, `list`, `hex`, `status`
- Modern CLI with progress spinners, ETA calculation, and colour themes
- Shell autocompletion generation (bash, zsh, fish, PowerShell)
- JSON output format support
- Hexadecimal result display option

#### Server Mode

- Production-ready REST API server
- Endpoints: `/calculate`, `/health`, `/algorithms`, `/metrics`
- Per-IP rate limiting (10 req/s, burst of 20)
- Security headers (X-Content-Type-Options, X-Frame-Options, CSP, etc.)
- Graceful shutdown with configurable timeout
- Request logging and metrics collection

#### Deployment

- Multi-stage Dockerfile with Alpine base (~15 MB image)
- Docker Compose configurations with monitoring stack
- Kubernetes manifests (Deployment, Service, HPA, PDB, NetworkPolicy)
- Helm chart support
- Non-root container execution

#### Documentation

- Comprehensive README with production deployment guide
- Architecture documentation with ADRs
- Performance tuning guide
- Security policy with vulnerability disclosure process
- REST API documentation with OpenAPI 3.0 specification
- Postman collection for API testing
- Algorithm-specific documentation (Fast Doubling, Matrix, FFT, GMP)
- Docker and Kubernetes deployment guides

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
