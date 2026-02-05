# FibGoIng: Refactor, Optimize & Finalize — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Clean dead code, fix quality issues, recreate the CLI entry point, and align all documentation to produce a finalized, buildable project.

**Architecture:** The project is a CLI-only high-performance Fibonacci calculator in Go. The `internal/app` package is the orchestrator; we need to recreate `cmd/fibcalc/main.go` (~25 lines) that wires `app.New()` and `app.Run()`. Dead code from the prior server/TUI/REPL refactor must be removed. Config field `Concise` must be renamed to `ShowValue` for clarity.

**Tech Stack:** Go 1.25+, `golang.org/x/sync/errgroup`, `github.com/briandowns/spinner`, `github.com/rs/zerolog`

---

## Task 1: Remove dead error types from `internal/errors`

**Files:**
- Modify: `internal/errors/errors.go:75-174`
- Modify: `internal/errors/errors_test.go` (remove tests for deleted types)

**Step 1: Delete ServerError type, methods, and constructor**

Remove lines 75-112 from `internal/errors/errors.go`:
- `ServerError` struct (lines 75-82)
- `ServerError.Error()` method (lines 84-94)
- `ServerError.Unwrap()` method (lines 96-100)
- `NewServerError()` function (lines 102-112)

**Step 2: Delete ValidationError type, methods, and constructor**

Remove lines 144-174 from `internal/errors/errors.go`:
- `ValidationError` struct (lines 144-153)
- `ValidationError.Error()` method (lines 155-161)
- `NewValidationError()` function (lines 163-174)

**Step 3: Remove tests for deleted types**

In `internal/errors/errors_test.go`, remove any test functions that reference `ServerError`, `NewServerError`, `ValidationError`, or `NewValidationError`.

**Step 4: Run tests to verify nothing breaks**

Run: `go test -v ./internal/errors/...`
Expected: PASS (remaining ConfigError, CalculationError, WrapError, IsContextError tests still pass)

**Step 5: Commit**

```bash
git add internal/errors/errors.go internal/errors/errors_test.go
git commit -m "refactor(errors): remove dead ServerError and ValidationError types"
```

---

## Task 2: Remove dead lifecycle functions from `internal/app`

**Files:**
- Delete: `internal/app/lifecycle.go` (entire file — all 4 exports are unused in production)
- Modify: `internal/app/lifecycle_test.go` (delete file — tests only tested dead code)

**Step 1: Delete `internal/app/lifecycle.go`**

The entire file contains only dead code:
- `SetupContext()` (line 20) — wrapper around `context.WithTimeout()`, never called
- `SetupSignals()` (line 33) — wrapper around `signal.NotifyContext()`, never called
- `SetupLifecycle()` (line 48) — combines above, never called
- `CancelFuncs` type + `Cleanup()` (line 58-76) — companion to above, never used

`app.go:177-180` already inlines this logic directly.

**Step 2: Delete `internal/app/lifecycle_test.go`**

Tests only exercise the deleted functions.

**Step 3: Run tests to verify nothing breaks**

Run: `go test -v ./internal/app/...`
Expected: PASS (remaining app tests don't depend on lifecycle.go)

**Step 4: Commit**

```bash
git add -A internal/app/lifecycle.go internal/app/lifecycle_test.go
git commit -m "refactor(app): remove unused lifecycle helper functions"
```

---

## Task 3: Remove dead code from `internal/calibration`

**Files:**
- Modify: `internal/calibration/adaptive.go`
- Modify: `internal/calibration/profile.go`
- Modify: `internal/calibration/microbench.go`
- Modify: corresponding `*_test.go` files (remove tests for deleted functions)

**Step 1: Clean `adaptive.go` — remove unused threshold generators**

Delete these functions (all confirmed never called in production):
- `GenerateFFTThresholds()` (lines 85-113)
- `GenerateStrassenThresholds()` (lines 127-146)
- `ValidateThresholds()` (lines 205-232)
- `GenerateFullThresholdSet()` (lines 245-252)
- `GenerateQuickThresholdSet()` (lines 254-261)
- `EstimatedThresholds()` (lines 263-268)
- `SortThresholds()` method (lines 271-275)

Keep: `GenerateParallelThresholds()` (used by `RunCalibration()`), all `GenerateQuick*()` functions, all `EstimateOptimal*()` functions.

If `ThresholdSet` type is only used by deleted functions, delete it too.

**Step 2: Clean `profile.go` — remove N-range optimization dead code**

Delete:
- `RangeThresholds` type (lines 43-60)
- `DefaultNRanges` variable (if present, companion to RangeThresholds)
- `GetThresholdsForN()` method (lines 219-242)
- `AddRangeThresholds()` method (lines 244-268)
- `InitializeDefaultRanges()` method (lines 270-288)
- `ProfileExists()` function (lines 307-314)
- Make `LoadProfile()` unexported (rename to `loadProfile()`) — only called by `LoadOrCreateProfile()`

Also remove the `ThresholdsByRange` field from `CalibrationProfile` struct if it references the deleted `RangeThresholds` type.

**Step 3: Clean `microbench.go` — remove unused wrapper**

Delete: `QuickCalibrateWithDefaults()` (lines 366-383) — `QuickCalibrate()` is used directly instead.

**Step 4: Clean test files**

Remove test functions that exercise deleted code in:
- `internal/calibration/adaptive_test.go`
- `internal/calibration/profile_test.go`
- `internal/calibration/microbench_test.go`

**Step 5: Run tests**

Run: `go test -v ./internal/calibration/...`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/calibration/
git commit -m "refactor(calibration): remove 200+ lines of dead code and incomplete N-range design"
```

---

## Task 4: Remove dead code from `internal/fibonacci`

**Files:**
- Modify: `internal/fibonacci/matrix_types.go`
- Modify: `internal/fibonacci/common.go`
- Modify: `internal/fibonacci/matrix_framework.go`

**Step 1: Remove unused `s9`, `s10` fields from `matrixState`**

In `internal/fibonacci/matrix_types.go`:
- Remove `s9, s10` from the struct field declaration (line 81)
- Remove their allocation in the pool `New` func (lines 120-121)
- Remove their nil checks in `releaseMatrixState()` (lines 164-165)

**Step 2: Remove unused `karatsubaThreshold` from task structs**

In `internal/fibonacci/common.go`:
- Remove `karatsubaThreshold int` field from `multiplicationTask` struct (line 55)
- Remove `karatsubaThreshold int` field from `squaringTask` struct (line 73)
- Find where these fields are set (likely in `doubling_framework.go`) and remove those assignments

**Step 3: Remove unused `multiplyMatricesFunc` variable**

In `internal/fibonacci/matrix_framework.go`:
- If `multiplyMatricesFunc` (line 20) is never overridden in tests and the function can be called directly, remove the variable and call `multiplyMatrices()` directly at the call site (line 68).
- Keep `squareSymmetricMatrixFunc` if it IS used.

**Step 4: Run tests**

Run: `go test -v -race ./internal/fibonacci/...`
Expected: PASS

**Step 5: Run benchmarks to check for regressions**

Run: `go test -bench=BenchmarkFastDoubling -benchmem ./internal/fibonacci/`
Expected: No performance regression (changes are removing unused fields)

**Step 6: Commit**

```bash
git add internal/fibonacci/
git commit -m "refactor(fibonacci): remove unused struct fields and variables"
```

---

## Task 5: Remove `writeOut()` wrapper from CLI

**Files:**
- Modify: `internal/cli/calculate.go:48-84`

**Step 1: Replace all `writeOut()` calls with `fmt.Fprintf()`**

In `internal/cli/calculate.go`, replace the 6 calls to `writeOut(out, ...)` at lines 48-71 with direct `fmt.Fprintf(out, ...)` calls.

**Step 2: Delete `writeOut()` function**

Remove lines 74-84 (the function and its doc comment).

**Step 3: Run tests**

Run: `go test -v ./internal/cli/...`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/cli/calculate.go
git commit -m "refactor(cli): remove writeOut wrapper, use fmt.Fprintf directly"
```

---

## Task 6: Rename `Concise` to `ShowValue` across the codebase

**Files:**
- Modify: `internal/config/config.go:88-90` (field rename)
- Modify: `internal/config/config.go:185-186` (flag definition)
- Modify: `internal/config/env.go` (if env override exists for Concise)
- Modify: `internal/cli/output.go:41-42` (OutputConfig field)
- Modify: `internal/cli/presenter.go:93` (parameter name in PresentResult)
- Modify: `internal/orchestration/interfaces.go:74` (ResultPresenter.PresentResult parameter name)
- Modify: `internal/orchestration/orchestrator.go` (AnalyzeComparisonResults call)
- Modify: `internal/app/app.go:215,237` (field assignment)
- Modify: `internal/cli/ui.go` (DisplayResult parameter name)
- Modify: all corresponding test files

**Step 1: Rename field in `AppConfig`**

In `internal/config/config.go`:
- Line 88-90: Change `Concise bool` to `ShowValue bool`
- Update the comment to: `// ShowValue, if true, displays the calculated Fibonacci value. Set with -c/--calculate.`

**Step 2: Update flag definitions**

In `internal/config/config.go`:
- Line 185: `fs.BoolVar(&config.ShowValue, "calculate", false, "Display the calculated value.")`
- Line 186: `fs.BoolVar(&config.ShowValue, "c", false, "Display the calculated value (shorthand).")`

**Step 3: Rename in `OutputConfig`**

In `internal/cli/output.go`:
- Line 42: Change `Concise bool` to `ShowValue bool`
- Update comment

**Step 4: Rename in `ResultPresenter` interface**

In `internal/orchestration/interfaces.go`:
- Line 76: Change parameter name `concise` to `showValue` in `PresentResult` signature

**Step 5: Propagate rename through all callers**

Update all references from `.Concise` to `.ShowValue` and parameter names from `concise` to `showValue` in:
- `internal/app/app.go:215` (`Concise: a.Config.Concise` → `ShowValue: a.Config.ShowValue`)
- `internal/app/app.go:237` (if Concise passed directly)
- `internal/cli/presenter.go:93` (PresentResult parameter)
- `internal/cli/ui.go` (DisplayResult parameter)
- `internal/orchestration/orchestrator.go` (AnalyzeComparisonResults if it passes concise)

**Step 6: Update all test files**

Search and replace `Concise` → `ShowValue` and `concise` → `showValue` in all `_test.go` files across the affected packages.

**Step 7: Run full test suite**

Run: `go test -v -race ./...`
Expected: PASS (compile errors if any reference was missed)

**Step 8: Commit**

```bash
git add internal/config/ internal/cli/ internal/orchestration/ internal/app/
git commit -m "refactor(config): rename Concise to ShowValue for semantic clarity"
```

---

## Task 7: Fix stale comment in `version.go`

**Files:**
- Modify: `internal/app/version.go:27`

**Step 1: Update the comment**

Change line 27 from:
```go
// This allows --version to work in any position (e.g., "fibcalc --server --version").
```
To:
```go
// This allows --version to work in any position (e.g., "fibcalc -n 100000 --version").
```

**Step 2: Commit**

```bash
git add internal/app/version.go
git commit -m "docs(app): fix stale --server reference in version.go comment"
```

---

## Task 8: Create `cmd/fibcalc/main.go` entry point

**Files:**
- Create: `cmd/fibcalc/main.go`

**Step 1: Create the entry point file**

Create `cmd/fibcalc/main.go`:

```go
package main

import (
	"context"
	"os"

	"github.com/agbru/fibcalc/internal/app"
)

func main() {
	if app.HasVersionFlag(os.Args) {
		app.PrintVersion(os.Stdout)
		return
	}

	application, err := app.New(os.Args, os.Stderr)
	if err != nil {
		if !app.IsHelpError(err) {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		os.Exit(1)
	}

	exitCode := application.Run(context.Background(), os.Stdout)
	os.Exit(exitCode)
}
```

Note: Version check is done BEFORE `app.New()` so `--version` works without valid config.

**Step 2: Verify it compiles**

Run: `go build -o build/fibcalc ./cmd/fibcalc`
Expected: Binary created at `build/fibcalc`

**Step 3: Test basic CLI paths**

Run these commands and verify non-error output:
```bash
go run ./cmd/fibcalc --version
go run ./cmd/fibcalc --help
go run ./cmd/fibcalc -n 100 -algo fast
go run ./cmd/fibcalc -n 1000 -algo all
go run ./cmd/fibcalc -n 100 --json
go run ./cmd/fibcalc -n 100 -q
go run ./cmd/fibcalc -n 100 --hex
```

**Step 4: Commit**

```bash
git add cmd/fibcalc/main.go
git commit -m "feat(cli): recreate cmd/fibcalc entry point"
```

---

## Task 9: Fix Makefile LDFLAGS and build targets

**Files:**
- Modify: `Makefile:22-25`

**Step 1: Fix LDFLAGS to target correct package**

The version variables are in `internal/app/version.go` (package `app`), not `main`. Change lines 22-25:

From:
```makefile
LDFLAGS=-ldflags="-s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)"
```

To:
```makefile
LDFLAGS=-ldflags="-s -w \
	-X github.com/agbru/fibcalc/internal/app.Version=$(VERSION) \
	-X github.com/agbru/fibcalc/internal/app.Commit=$(COMMIT) \
	-X github.com/agbru/fibcalc/internal/app.BuildDate=$(BUILD_DATE)"
```

**Step 2: Make PGO build conditional**

Find the `build` target and make PGO optional — if `$(PGO_PROFILE)` doesn't exist, build without PGO. The build target should work whether or not the PGO profile file exists.

**Step 3: Verify build targets work**

Run (on a system with make):
```bash
make build
make test
make lint
```

Or without make:
```bash
go build -ldflags="-s -w -X github.com/agbru/fibcalc/internal/app.Version=test" -o build/fibcalc ./cmd/fibcalc
./build/fibcalc --version
```
Expected: Shows "fibcalc test" with commit and date info.

**Step 4: Clean dependencies**

Run: `go mod tidy && go mod verify`
Expected: go.sum cleaned of stale transitive entries.

**Step 5: Commit**

```bash
git add Makefile go.sum
git commit -m "fix(build): fix LDFLAGS package path and make PGO conditional"
```

---

## Task 10: Run full validation suite

**Files:** None modified — validation only.

**Step 1: Full test suite with race detector**

Run: `go test -v -race -cover ./...`
Expected: All PASS, coverage >75%.

**Step 2: Cross-platform build check**

```bash
GOOS=linux GOARCH=amd64 go build ./cmd/fibcalc
GOOS=windows GOARCH=amd64 go build ./cmd/fibcalc
GOOS=darwin GOARCH=arm64 go build ./cmd/fibcalc
```
Expected: All compile without errors.

**Step 3: Linting**

Run: `golangci-lint run ./...`
Expected: No errors (warnings acceptable).

**Step 4: Benchmark regression check**

Run: `go test -bench=. -benchmem ./internal/fibonacci/ -short`
Expected: No significant regression vs. pre-refactor.

**Step 5: Regenerate mocks**

Run: `go generate ./...`
Then: `go test -v ./...` to verify regenerated mocks work.

**Step 6: Commit if mocks changed**

```bash
git add internal/fibonacci/mocks/ internal/cli/mocks/
git commit -m "chore: regenerate mocks after interface changes"
```

---

## Task 11: Update documentation

**Files:**
- Modify: `README.md` (fix 4 broken sections)
- Modify: `CONTRIBUTING.md` (fix 2 stale build examples)
- Modify: `Docs/algorithms/FAST_DOUBLING.md` (3 stale CLI examples)
- Modify: `Docs/algorithms/FFT.md` (9 stale CLI examples)
- Modify: `Docs/algorithms/GMP.md` (2 stale CLI examples)
- Modify: `Docs/algorithms/MATRIX.md` (2 stale CLI examples)
- Modify: `Docs/PERFORMANCE.md` (2 stale build commands)
- Modify: `Docs/ARCHITECTURE.md` (minor PGO path update)
- Modify: `CLAUDE.md` (remove known issue about missing entry point, update field name)

**Step 1: Fix README.md**

- Quick Start: Change `go run ./cmd/fibcalc -n 10000000` (confirm it works after Task 8)
- Installation: `go install github.com/agbru/fibcalc/cmd/fibcalc@latest`
- Build: Confirm `make build` works after Task 9
- Remove or create `Docs/TROUBLESHOOTING.md` reference (line ~260). Prefer removing the reference since the file doesn't exist.

**Step 2: Fix all `./fibcalc` references in Docs/algorithms/**

Replace `./fibcalc` with `go run ./cmd/fibcalc` or `fibcalc` (assuming installed) in:
- `FAST_DOUBLING.md` (3 instances)
- `FFT.md` (9 instances)
- `GMP.md` (2 instances)
- `MATRIX.md` (2 instances)

**Step 3: Fix CONTRIBUTING.md**

Replace stale build commands with working ones.

**Step 4: Fix PERFORMANCE.md**

Replace stale `go build ... ./cmd/fibcalc` examples.

**Step 5: Update ARCHITECTURE.md**

- Confirm `cmd/fibcalc` is listed as present (not "removed")
- Update PGO profile path if needed

**Step 6: Update CLAUDE.md**

- Remove "Known Issues" bullet about `cmd/fibcalc` not existing
- Change `Concise` → `ShowValue` in any interface/field documentation
- Update any stale file references

**Step 7: Commit**

```bash
git add README.md CONTRIBUTING.md CLAUDE.md Docs/
git commit -m "docs: update all documentation after refactor and CLI restoration"
```

---

## Task 12: Update project memory

**Files:**
- Modify: `~/.claude/projects/.../memory/MEMORY.md`

**Step 1: Update MEMORY.md**

Reflect the new project state:
- `cmd/fibcalc` EXISTS now
- Dead code removed from errors, app, calibration, fibonacci
- `Concise` renamed to `ShowValue`
- Makefile LDFLAGS fixed
- Documentation aligned

---

## Summary

| Task | Description | Type |
|------|-------------|------|
| 1 | Remove dead error types (ServerError, ValidationError) | Dead code removal |
| 2 | Remove dead lifecycle functions | Dead code removal |
| 3 | Remove dead calibration code (~200 lines) | Dead code removal |
| 4 | Remove dead fibonacci struct fields | Dead code removal |
| 5 | Remove writeOut() wrapper | Dead code removal |
| 6 | Rename Concise → ShowValue | Quality refactor |
| 7 | Fix stale --server comment | Docs fix |
| 8 | Create cmd/fibcalc/main.go | Feature |
| 9 | Fix Makefile LDFLAGS + clean deps | Build fix |
| 10 | Full validation suite | Testing |
| 11 | Update all documentation | Docs |
| 12 | Update project memory | Housekeeping |

**Total: 12 tasks, 12 commits, each independently testable.**
