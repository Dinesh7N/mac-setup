package installer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"macsetup/internal/config"
	"macsetup/internal/utils"
)

type Manager struct {
	maxWorkers int
	progress   chan ProgressUpdate
	verbose    bool
}

func NewManager(maxWorkers int, opts RunOptions) *Manager {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	return &Manager{
		maxWorkers: maxWorkers,
		progress:   make(chan ProgressUpdate, 128),
		verbose:    opts.Verbose,
	}
}

func (m *Manager) Progress() <-chan ProgressUpdate {
	return m.progress
}

func (m *Manager) Run(ctx context.Context, selected map[string]bool) (Summary, error) {
	defer close(m.progress)

	pkgs := selectedPackages(selected)
	results := make([]InstallResult, 0, len(pkgs)+10)

	if !IsXcodeInstalled(ctx) {
		task := config.Package{Name: "Xcode CLI Tools", Type: config.TypeSystem, Category: "core", Required: true, Default: true}
		m.emit(task, StatusRunning, "", "")
		status, msg, errStr, dur := timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
			_ = TriggerXcodeInstall(ctx)
			if err := WaitForXcode(ctx, 2*time.Second); err != nil {
				return StatusFailed, "", err
			}
			return StatusInstalled, "", nil
		})
		if errStr != "" {
			errStr = classifyInstallError(task, fmt.Errorf("%s", errStr))
		}
		m.emit(task, status, msg, errStr)
		results = append(results, InstallResult{Package: task, Status: status, Message: msg, Error: errStr, Duration: dur})
	} else {
		task := config.Package{Name: "Xcode CLI Tools", Type: config.TypeSystem, Category: "core", Required: true, Default: true}
		m.emit(task, StatusSkipped, "Already installed", "")
		results = append(results, InstallResult{Package: task, Status: StatusSkipped, Message: "Already installed"})
	}

	if !IsBrewInstalled(ctx, m.verbose) {
		task := config.Package{Name: "Homebrew", Type: config.TypeSystem, Category: "core", Required: true, Default: true}
		m.emit(task, StatusRunning, "", "")
		status, msg, errStr, dur := timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
			if err := InstallBrew(ctx, m.verbose); err != nil {
				return StatusFailed, "", err
			}
			return StatusInstalled, "", nil
		})
		if errStr != "" {
			errStr = classifyInstallError(task, fmt.Errorf("%s", errStr))
		}
		m.emit(task, status, msg, errStr)
		results = append(results, InstallResult{Package: task, Status: status, Message: msg, Error: errStr, Duration: dur})
	} else {
		task := config.Package{Name: "Homebrew", Type: config.TypeSystem, Category: "core", Required: true, Default: true}
		m.emit(task, StatusSkipped, "Already installed", "")
		results = append(results, InstallResult{Package: task, Status: StatusSkipped, Message: "Already installed"})
	}

	{
		task := config.Package{Name: "Homebrew update", Type: config.TypeTask, Category: "core", Required: true, Default: true}
		m.emit(task, StatusRunning, "", "")
		status, msg, errStr, dur := timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
			if err := BrewUpdate(ctx, m.verbose); err != nil {
				return StatusSkipped, "Update failed (non-critical)", nil
			}
			if err := BrewUpgrade(ctx, m.verbose); err != nil {
				return StatusFailed, "", err
			}
			return StatusInstalled, "", nil
		})
		if errStr != "" {
			errStr = classifyInstallError(task, fmt.Errorf("%s", errStr))
		}
		m.emit(task, status, msg, errStr)
		results = append(results, InstallResult{Package: task, Status: status, Message: msg, Error: errStr, Duration: dur})
	}

	taps, formulas, casks := splitBrewPackages(pkgs)

	for _, tap := range taps {
		m.emit(tap, StatusRunning, "", "")
		status, msg, errStr, dur := timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
			installed, err := IsTapInstalled(ctx, m.verbose, tap.Tap)
			if err != nil {
				return StatusFailed, "", err
			}
			if installed {
				return StatusSkipped, "Already tapped", nil
			}
			if err := AddTap(ctx, m.verbose, tap.Tap); err != nil {
				return StatusFailed, "", err
			}
			return StatusInstalled, "", nil
		})
		if errStr != "" {
			errStr = classifyInstallError(tap, fmt.Errorf("%s", errStr))
		}
		m.emit(tap, status, msg, errStr)
		results = append(results, InstallResult{Package: tap, Status: status, Message: msg, Error: errStr, Duration: dur})
	}

	results = append(results, m.installFormulas(ctx, formulas)...)
	for _, cask := range casks {
		m.emit(cask, StatusRunning, "", "")
		status, msg, errStr, dur := timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
			// First check if installed via Homebrew
			installed, err := IsBrewPackageInstalled(ctx, m.verbose, cask)
			if err != nil {
				return StatusFailed, "", err
			}
			if installed {
				return StatusSkipped, "Already installed", nil
			}

			// Check if app exists manually in /Applications
			appExists, appPath := IsCaskAppInstalled(cask.Name)
			if appExists {
				return StatusSkipped, fmt.Sprintf("Already installed at %s", appPath), nil
			}

			if err := InstallCask(ctx, m.verbose, cask.Name); err != nil {
				return StatusFailed, "", err
			}
			return StatusInstalled, "", nil
		})
		if errStr != "" {
			errStr = classifyInstallError(cask, fmt.Errorf("%s", errStr))
		}
		m.emit(cask, status, msg, errStr)
		results = append(results, InstallResult{Package: cask, Status: status, Message: msg, Error: errStr, Duration: dur})
	}

	postResults := m.postInstall(ctx)
	results = append(results, postResults...)

	{
		task := config.Package{Name: "Post-install verification", Type: config.TypeTask, Category: "core", Required: true, Default: true}
		m.emit(task, StatusRunning, "", "")
		start := time.Now()
		ver := VerifyCriticalTools(ctx)
		_, failed, summaryMsg := VerifySummary(ver)
		status := StatusInstalled
		errStr := ""
		if failed > 0 {
			status = StatusFailed
			errStr = summaryMsg
		}
		m.emit(task, status, "", errStr)
		results = append(results, InstallResult{Package: task, Status: status, Error: errStr, Duration: time.Since(start)})
		if failed > 0 {
			for _, r := range ver {
				if r.Error == "" {
					continue
				}
				p := config.Package{Name: "verify: " + r.Name, Type: config.TypeTask, Category: "core"}
				results = append(results, InstallResult{Package: p, Status: StatusFailed, Error: r.Error})
			}
		}
	}

	return Summary{Results: results}, nil
}

func (m *Manager) postInstall(ctx context.Context) []InstallResult {
	var results []InstallResult

	dirsTask := config.Package{Name: "Create config directories", Type: config.TypeTask, Category: "core"}
	m.emit(dirsTask, StatusRunning, "", "")
	st, msg, errStr, dur := timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		if err := CreateConfigDirectories(); err != nil {
			return StatusFailed, "", err
		}
		return StatusInstalled, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(dirsTask, fmt.Errorf("%s", errStr))
	}
	m.emit(dirsTask, st, msg, errStr)
	results = append(results, InstallResult{Package: dirsTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	ohTask := config.Package{Name: "Oh My Zsh", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(ohTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		installed, err := IsOhMyZshInstalled()
		if err != nil {
			return StatusFailed, "", err
		}
		if installed {
			return StatusSkipped, "Already installed", nil
		}
		if err := InstallOhMyZsh(ctx); err != nil {
			return StatusFailed, "", err
		}
		return StatusInstalled, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(ohTask, fmt.Errorf("%s", errStr))
	}
	m.emit(ohTask, st, msg, errStr)
	results = append(results, InstallResult{Package: ohTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	pluginsTask := config.Package{Name: "Zsh plugins", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(pluginsTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		outcome, err := InstallZshPlugins(ctx)
		if err != nil {
			return StatusFailed, "", err
		}
		if outcome == StatusSkipped {
			return StatusSkipped, "Already installed", nil
		}
		return StatusInstalled, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(pluginsTask, fmt.Errorf("%s", errStr))
	}
	m.emit(pluginsTask, st, msg, errStr)
	results = append(results, InstallResult{Package: pluginsTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	nvimTask := config.Package{Name: "Neovim config (kickstart)", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(nvimTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		outcome, err := CloneNeovimConfig(ctx)
		if err != nil {
			return StatusFailed, "", err
		}
		return outcome, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(nvimTask, fmt.Errorf("%s", errStr))
	}
	m.emit(nvimTask, st, msg, errStr)
	results = append(results, InstallResult{Package: nvimTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	tpmTask := config.Package{Name: "tmux plugin manager (TPM)", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(tpmTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		outcome, err := CloneTPM(ctx)
		if err != nil {
			return StatusFailed, "", err
		}
		return outcome, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(tpmTask, fmt.Errorf("%s", errStr))
	}
	m.emit(tpmTask, st, msg, errStr)
	results = append(results, InstallResult{Package: tpmTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	miseTask := config.Package{Name: "Mise runtimes", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(miseTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		if _, err := utils.Run(ctx, m.verbose, 5*time.Second, "mise", "--version"); err != nil {
			return StatusSkipped, "mise not installed yet", nil
		}
		if err := SetupMise(ctx); err != nil {
			return StatusFailed, "", err
		}
		return StatusInstalled, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(miseTask, fmt.Errorf("%s", errStr))
	}
	m.emit(miseTask, st, msg, errStr)
	results = append(results, InstallResult{Package: miseTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	dotTask := config.Package{Name: "Dotfiles", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(dotTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		backupCount, err := WriteDotfiles()
		if err != nil {
			return StatusFailed, "", err
		}
		if backupCount > 0 {
			return StatusInstalled, fmt.Sprintf("Backed up %d file(s)", backupCount), nil
		}
		return StatusInstalled, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(dotTask, fmt.Errorf("%s", errStr))
	}
	m.emit(dotTask, st, msg, errStr)
	results = append(results, InstallResult{Package: dotTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	fzfTask := config.Package{Name: "Configure fzf", Type: config.TypeTask, Category: "shell_cli"}
	m.emit(fzfTask, StatusRunning, "", "")
	st, msg, errStr, dur = timed(ctx, m.verbose, func(ctx context.Context) (InstallStatus, string, error) {
		outcome, err := ConfigureFzf(ctx)
		if err != nil {
			return StatusFailed, "", err
		}
		return outcome, "", nil
	})
	if errStr != "" {
		errStr = classifyInstallError(fzfTask, fmt.Errorf("%s", errStr))
	}
	m.emit(fzfTask, st, msg, errStr)
	results = append(results, InstallResult{Package: fzfTask, Status: st, Message: msg, Error: errStr, Duration: dur})

	return results
}

func (m *Manager) installFormulas(ctx context.Context, formulas []config.Package) []InstallResult {
	jobs := make(chan config.Package)
	out := make(chan InstallResult, len(formulas))
	var wg sync.WaitGroup

	workers := m.maxWorkers
	if workers <= 0 {
		workers = 5
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pkg := range jobs {
				m.emit(pkg, StatusRunning, "", "")

				start := time.Now()
				status := StatusInstalled
				msg := ""
				errStr := ""

				installed, err := IsBrewPackageInstalled(ctx, m.verbose, pkg)
				if err != nil {
					status = StatusFailed
					errStr = err.Error()
				} else if installed {
					// Package is installed, but check if it's linked
					if !IsFormulaLinked(ctx, m.verbose, pkg.Name) {
						// Try to link it
						if err := LinkFormula(ctx, m.verbose, pkg.Name); err != nil {
							status = StatusFailed
							errStr = fmt.Sprintf("installed but not linked: %v", err)
						} else {
							status = StatusSkipped
							msg = "Already installed (relinked)"
						}
					} else {
						status = StatusSkipped
						msg = "Already installed"
					}
				} else {
					if err := InstallFormula(ctx, m.verbose, pkg.Name); err != nil {
						status = StatusFailed
						errStr = classifyInstallError(pkg, err)
					}
				}

				m.emit(pkg, status, msg, errStr)
				out <- InstallResult{Package: pkg, Status: status, Message: msg, Error: errStr, Duration: time.Since(start)}
			}
		}()
	}

	go func() {
		for _, f := range formulas {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- f:
			}
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()

	var results []InstallResult
	for r := range out {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool {
		return strings.Compare(results[i].Package.Name, results[j].Package.Name) < 0
	})
	return results
}

func (m *Manager) emit(pkg config.Package, status InstallStatus, message, errStr string) {
	select {
	case m.progress <- ProgressUpdate{Package: pkg, Status: status, Message: message, Error: errStr}:
	default:
	}
}

func splitBrewPackages(pkgs []config.Package) (taps []config.Package, formulas []config.Package, casks []config.Package) {
	seenTap := make(map[string]bool)
	for _, pkg := range pkgs {
		switch pkg.Type {
		case config.TypeTap:
			if pkg.Tap != "" && !seenTap[pkg.Tap] {
				seenTap[pkg.Tap] = true
				taps = append(taps, pkg)
			}
		case config.TypeFormula:
			if pkg.Tap != "" && !seenTap[pkg.Tap] {
				seenTap[pkg.Tap] = true
				taps = append(taps, config.Package{Name: pkg.Tap, Type: config.TypeTap, Category: pkg.Category, Tap: pkg.Tap, Default: pkg.Default})
			}
			formulas = append(formulas, pkg)
		case config.TypeCask:
			casks = append(casks, pkg)
		}
	}
	sort.Slice(taps, func(i, j int) bool { return strings.Compare(taps[i].Tap, taps[j].Tap) < 0 })
	sort.Slice(formulas, func(i, j int) bool { return strings.Compare(formulas[i].Name, formulas[j].Name) < 0 })
	sort.Slice(casks, func(i, j int) bool { return strings.Compare(casks[i].Name, casks[j].Name) < 0 })
	return taps, formulas, casks
}
