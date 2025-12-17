# Code Review Feedback

**Rating:** 9/10

The codebase is high-quality, idiomatic Go, and adheres strictly to the provided `IMPLEMENTATION_SPEC.md`. The separation of concerns between configuration, installation logic, and TUI is clean and maintainable. The use of the `bubbletea` framework is effective, and the concurrency model for package installation is well-implemented.

## Strengths

1.  **Adherence to Spec:** The implementation faithfully follows the detailed specification, including package taxonomies, TUI states, and idempotency rules.
2.  **Concurrency:** The use of a worker pool and buffered channels for parallel Homebrew formula installation (`internal/installer/manager.go`) is efficient and correct.
3.  **Idempotency:** Robust checks are in place (e.g., `IsBrewPackageInstalled`, `WriteWithBackup`) to ensure the tool can be run multiple times safely.
4.  **TUI Design:** The TUI state machine in `internal/tui/app.go` is well-structured, handling user input and screen transitions smoothly.
5.  **Context Management:** Proper use of `context.Context` ensures operations can be timed out or cancelled.

## Suggestions for Improvement

### 1. Robustness & Error Handling

*   **Retry Mechanism:** Network operations (Homebrew installs, `curl`, `git clone`) can fail transiently. Consider adding a simple retry utility (e.g., "retry 3 times with exponential backoff") in `internal/utils` for these specific operations.
*   **Atomic File Writing:** In `internal/utils/filesystem.go`, `WriteWithBackup` renames the existing file to a backup *before* attempting to write the new file. If writing the new file fails (e.g., disk full), the user is left with no config file (only the backup).
    *   *Suggestion:* Write the new content to a temporary file first, then perform the backup rename, and finally rename the temp file to the destination.
*   **Sudo Keep-Alive:** The `KeepSudoAlive` goroutine runs indefinitely. While harmless in a CLI tool that exits, it's good practice to tie it to a `context.Context` to stop the ticker when the application finishes.

### 2. User Experience

*   **Headless Logging:** When running in `--headless` mode, output is printed to `os.Stdout`. It might be beneficial to support writing detailed logs (including `stderr` from failed commands) to a file (e.g., `macsetup.log`) for debugging purposes.
*   **Detailed Failure Info:** The summary screen shows that packages failed. Allowing the user to expand or view the full error message/stderr for failed packages within the TUI would be helpful.

### 3. Minor Edge Cases

*   **Backup Timestamp Collision:** `WriteWithBackup` uses a second-precision timestamp (`20060102_150405`). If the tool is run twice rapidly in automation or by accident, it could overwrite a backup.
    *   *Suggestion:* Append milliseconds or a random suffix to the backup filename.
*   **Intel Macs:** The `PreflightChecks` strictly enforce `arm64`. While the spec demands this, if you ever support Intel Macs, hardcoded paths like `/opt/homebrew` will need to be dynamic (`/usr/local` on Intel).

---

## Additional Suggestions (Code Review Round 2)

### 4. Testing & Quality Assurance

*   **Expand Unit Test Coverage:** Currently only `internal/config/packages_test.go` has 2 tests (`TestAllPackagesHaveCategory`, `TestNoDuplicatePackages`). Critical packages need test coverage:
    *   `internal/utils/filesystem.go` - Test `WriteWithBackup`, `ExpandHome`, `SymlinkIfMissing`
    *   `internal/utils/exec.go` - Test `RunCommand` with mock commands
    *   `internal/installer/homebrew.go` - Test `IsBrewPackageInstalled` parsing logic
    *   `internal/utils/errors.go` - Test `ClassifyError` with various stderr patterns

*   **Integration Tests:** Add integration tests that run against a mock or containerized environment to validate the full installation flow without affecting the host system.

*   **Add Linter Configuration:** The Makefile references a `lint` target, but `.golangci.yaml` is missing. Add a linter config to enforce code quality:
    ```yaml
    # .golangci.yaml
    linters:
      enable:
        - errcheck
        - govet
        - staticcheck
        - unused
        - gosimple
        - ineffassign
    ```

### 5. Code Structure & Organization

*   **Extract TUI Styles:** The IMPLEMENTATION_SPEC.md mentions a separate `internal/tui/styles.go`, but all styles are currently inline in `app.go` (489 lines). Extracting Lip Gloss styles to a dedicated file would improve maintainability.

*   **Unused Error Classification:** `internal/utils/errors.go` defines `ClassifyError()` with types (network, permission, dependency, etc.), but this isn't used by the installer. Integrate error classification into the summary view to show users *why* something failed (e.g., "Network error - check your connection").

*   **Constants Extraction:** Magic strings like Homebrew URLs, plugin repositories, and paths are scattered across installer files. Consider a `internal/constants/` package or constants file:
    ```go
    const (
        HomebrewInstallURL = "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh"
        OhMyZshInstallURL  = "https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh"
        // ...
    )
    ```

### 6. Feature Enhancements

*   **Dry-Run Mode:** Add a `--dry-run` flag that shows what would be installed without making changes. This is valuable for users who want to preview the setup:
    ```go
    // cmd/root.go
    rootCmd.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")
    ```

*   **Config File Support:** Allow users to provide a YAML/JSON config file to customize package selection, useful for team-wide defaults:
    ```bash
    macsetup --config team-setup.yaml
    ```

*   **Post-Install Verification:** After installation completes, verify that critical tools are actually available in PATH and working (e.g., `brew --version`, `git --version`, `nvim --version`).

*   **Resume/Continue Support:** If the tool crashes mid-installation, there's no way to resume. Consider writing state to a temp file and offering `--resume` functionality.

### 7. Project Maintenance

*   **Missing Project Files:** Add standard open-source project files:
    *   `LICENSE` - Choose an appropriate license (MIT, Apache 2.0, etc.)
    *   `CHANGELOG.md` - Track version changes
    *   `CONTRIBUTING.md` - Guide for contributors (if open-sourcing)

*   **Structured Logging:** Replace `fmt.Fprintf` calls with structured logging using `log/slog` (Go 1.21+) or `zerolog`. This enables:
    *   Log levels (debug, info, warn, error)
    *   JSON output for parsing
    *   File logging in headless mode

*   **Version Subcommand Enhancement:** The `version` command shows version/commit/date. Consider adding:
    *   Go version used to build
    *   Link to release notes
    *   Update check (compare with latest GitHub release)

### 8. Security Considerations

*   **Homebrew Script Verification:** The Homebrew install script is fetched via curl and piped to bash. Consider:
    *   Downloading to a temp file first
    *   Verifying checksum if Homebrew provides one
    *   Showing the user what will be executed

*   **Sudo Scope:** `KeepSudoAlive` runs `sudo -v` every 60 seconds. Document this behavior clearly so users understand why sudo is kept active.

### 9. Documentation

*   **README Improvements:** The README could include:
    *   GIF/screenshot of the TUI in action
    *   List of all packages installed by category
    *   Troubleshooting section for common issues
    *   Comparison with similar tools (e.g., thoughtbot/laptop)

*   **Code Comments:** Some complex functions lack doc comments:
    *   `Manager.Run()` in `manager.go` - Explain the installation phases
    *   `Update()` in `app.go` - Document the state transitions
    *   `WriteWithBackup()` - Document the backup naming convention

---

## Summary of Priorities

| Priority | Suggestion | Effort |
|----------|------------|--------|
| High | Expand unit test coverage | Medium |
| High | Add `.golangci.yaml` linter config | Low |
| High | Add `LICENSE` file | Low |
| Medium | Extract TUI styles to `styles.go` | Low |
| Medium | Add `--dry-run` flag | Medium |
| Medium | Integrate error classification in summary | Medium |
| ~~Medium~~ | ~~Add GitHub Actions CI workflow~~ | ~~Medium~~ | *Removed - see note below* |
| Low | Structured logging with slog | Medium |
| Low | Config file support | High |
| Low | Resume/continue functionality | High |

**Note on CI:** GitHub Actions CI was removed from recommendations. This tool only runs on macOS arm64, and the meaningful tests (Homebrew installs, dotfile deployment) can't run in CI. macOS runners are expensive/slow, and the unit tests can be run locally with `make test`. The practical approach is to run `make lint && make test` before commits and test on real Macs before releases.

---

## Code Review Round 3: Implementation Review

**Date:** Review of applied improvements

### Implemented Items - Assessment

#### 1. Retry with Backoff (`internal/utils/retry.go`) - Excellent

The implementation is clean and production-ready:
- Configurable attempts, base delay, and max delay
- Proper exponential backoff with jitter (using crypto/rand)
- Context-aware with proper cancellation handling
- Good defaults when options are zero

**Minor suggestions:**
- Consider adding a `ShouldRetry func(error) bool` option to allow callers to skip retries for non-transient errors (e.g., 404s shouldn't retry)
- Could log retry attempts for debugging (when structured logging is added)

#### 2. Atomic File Writing (`internal/utils/filesystem.go`) - Excellent

The atomic write pattern is now correct:
1. Write to temp file in same directory
2. Set permissions on temp file
3. Close temp file
4. Backup existing file (if exists)
5. Rename temp to destination
6. Rollback backup if rename fails

**Backup uniqueness** is also fixed with milliseconds + random hex suffix:
```go
timestamp := time.Now().Format("20060102_150405.000")
// + hex.EncodeToString(buf[:]) for uniqueness
```

**Test coverage** (`filesystem_test.go`) includes:
- `TestExpandHome` 
- `TestWriteWithBackup` 
- `TestWriteWithBackupUnique` (verifies distinct backups)
- `TestSymlinkIfMissing`

#### 3. Context-Aware Sudo Keep-Alive (`internal/utils/system.go`) - Correct

```go
func KeepSudoAlive(ctx context.Context) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            _ = exec.Command("sudo", "-n", "true").Run()
        }
    }
}
```

Properly stops when context is cancelled. Uses `sudo -n true` (non-interactive) which is correct.

#### 4. Dry-Run Mode (`internal/installer/dryrun.go`, `cmd/root.go`) - Good

The `--dry-run` / `-n` flag is properly implemented:
- Shows Xcode/Homebrew status (installed or would install)
- Lists all taps, formulas, and casks that would be installed
- Lists post-install tasks

**Suggestions for enhancement:**
- Show package descriptions alongside names
- Indicate which packages are already installed vs would be installed
- Add estimated time based on package count

#### 5. Headless File Logging (`cmd/root.go`) - Good

```go
if logFile != "" {
    f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
    // ...
    out = io.MultiWriter(os.Stdout, f)
}
```

Uses `io.MultiWriter` to write to both stdout and file. Proper cleanup with deferred close.

**Suggestion:** Consider adding timestamps to log entries for easier debugging.

#### 6. Error Classification Integration (`internal/installer/errors.go`, `manager.go`) - Correct

The `classifyInstallError()` wrapper properly integrates with the utils error classification:
```go
func classifyInstallError(pkg config.Package, err error) string {
    ie := utils.ClassifyError(pkg.Name, err, err.Error())
    return ie.Error()
}
```

Now used throughout `manager.go` for all failure cases, providing actionable error messages like:
- `[network] Network error - check your internet connection`
- `[permission] Permission denied - try running with sudo`
- `[not_found] Package not found in Homebrew`

#### 7. TUI Styles Extraction (`internal/tui/styles.go`) - Done

Styles are now in a separate file:
```go
var (
    titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
    dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
    badStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)
```

**Suggestion:** Consider adding more styles as the TUI grows:
- `selectedStyle` for highlighted items
- `categoryStyle` for category headers
- `helpStyle` for the help text at the bottom

#### 8. Linter Configuration (`.golangci.yaml`) - Good

Basic linters enabled:
- errcheck, govet, staticcheck, unused, gosimple, ineffassign

**Suggestion:** Consider adding more linters for stricter code quality:
```yaml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - gocritic      # Highly extensible Go linter
    - gofmt         # Checks if code is gofmt-ed
    - misspell      # Finds commonly misspelled English words
    - unconvert     # Remove unnecessary type conversions
    - prealloc      # Find slice declarations that could be preallocated
```

#### 9. Unit Tests (`internal/utils/*_test.go`) - Good Coverage

**filesystem_test.go:** 4 tests covering core functionality
**exec_test.go:** 2 tests (success case and timeout)
**errors_test.go:** Table-driven test with 4 cases (network, permission, not_found, dependency)

**Missing test coverage:**
- `retry.go` - Should test retry behavior, backoff timing, context cancellation
- `system.go` - Hard to test (requires mocking exec), but could test architecture detection

---

### Remaining Items from Previous Reviews

| Item | Status | Notes |
|------|--------|-------|
| Retry mechanism | **Done** | Well implemented with jitter |
| Atomic file writes | **Done** | Correct temp file swap pattern |
| Sudo context binding | **Done** | Properly stops on ctx.Done() |
| Dry-run mode | **Done** | Could show more detail |
| Headless logging | **Done** | Via `--log-file` flag |
| Error classification | **Done** | Integrated into manager |
| TUI styles extraction | **Done** | Basic styles extracted |
| Linter config | **Done** | Basic config added |
| Unit test expansion | **Done** | utils package covered |
| LICENSE file | Pending | |
| CHANGELOG.md | Pending | |
| ~~GitHub Actions CI~~ | **Removed** | Not practical for macOS-only tool |
| Config file support | Pending | Lower priority |
| Post-install verification | Pending | |
| Structured logging | Pending | Lower priority |
| Constants extraction | Pending | URLs still scattered |

---

### New Suggestions (Round 3)

#### 1. Add Retry Test Coverage

```go
// internal/utils/retry_test.go
func TestRetrySucceedsOnFirstAttempt(t *testing.T) { ... }
func TestRetrySucceedsAfterFailures(t *testing.T) { ... }
func TestRetryRespectsMaxAttempts(t *testing.T) { ... }
func TestRetryRespectsContext(t *testing.T) { ... }
func TestRetryBackoffTiming(t *testing.T) { ... }
```

#### 2. Improve Dry-Run Output

Currently shows package names only. Could show:
```
Dry run: planned steps
- Xcode CLI Tools: already installed (skip)
- Homebrew: already installed (skip install, would run brew update/upgrade)
- Formulas (parallel): 15
  - bat (modern cat replacement) - NOT INSTALLED
  - eza (modern ls replacement) - INSTALLED, would skip
  ...
```

#### 3. Add `--verbose` Flag

For debugging, add a verbose mode that shows:
- Full command output
- Retry attempts
- Timing information per package

#### 4. Consider Signal Handling

The application doesn't gracefully handle SIGINT/SIGTERM during installation. Consider:
```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()
```

This would allow graceful shutdown and proper cleanup.

#### 5. Validate Package Dependencies

Before installation, could check if packages have mutual dependencies and order them appropriately, or warn if a selected cask requires a formula that wasn't selected.

---

### Updated Rating: 9.2/10

The implementation quality has improved significantly. The retry mechanism, atomic writes, and error classification are all production-ready. Test coverage has expanded appropriately. The main remaining work is project hygiene (LICENSE, CI/CD) and optional enhancements.
