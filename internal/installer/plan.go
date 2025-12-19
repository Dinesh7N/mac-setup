package installer

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"macsetup/internal/config"
)

func DefaultSelection() map[string]bool {
	return config.DefaultSelection()
}

type PlanOptions struct {
	MaxWorkers int
}

type RunOptions struct {
	Verbose bool
}

func RunInstallPlan(ctx context.Context, selected map[string]bool, maxWorkers int, out io.Writer, opts RunOptions) (Summary, error) {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}

	manager := NewManager(maxWorkers, opts)
	updates := manager.Progress()
	done := make(chan Summary, 1)
	errCh := make(chan error, 1)

	go func() {
		summary, err := manager.Run(ctx, selected)
		if err != nil {
			errCh <- err
			return
		}
		done <- summary
	}()

	for updates != nil || done != nil || errCh != nil {
		select {
		case <-ctx.Done():
			return Summary{}, ctx.Err()
		case err := <-errCh:
			errCh = nil
			return Summary{}, err
		case summary := <-done:
			done = nil
			return summary, nil
		case upd, ok := <-updates:
			if !ok {
				updates = nil
				continue
			}
			line := formatProgress(upd, opts.Verbose)
			if out != nil {
				_, _ = fmt.Fprintln(out, line)
			}
		}
	}
	return Summary{}, fmt.Errorf("installation plan ended without summary")
}

func formatProgress(upd ProgressUpdate, verbose bool) string {
	var status string
	switch upd.Status {
	case StatusRunning:
		status = "..."
	case StatusInstalled:
		status = "installed"
	case StatusSkipped:
		status = "skipped"
	case StatusFailed:
		status = "failed"
	default:
		status = string(upd.Status)
	}

	name := upd.Package.Name
	if upd.Package.Type != config.TypeSystem && upd.Package.Type != config.TypeTask {
		name = fmt.Sprintf("%s (%s)", upd.Package.Name, upd.Package.Type)
	}
	if verbose && upd.Package.Category != "" {
		name = fmt.Sprintf("%s [%s]", name, upd.Package.Category)
	}
	if upd.Message != "" {
		return fmt.Sprintf("%s: %s - %s", name, status, upd.Message)
	}
	if upd.Error != "" {
		return fmt.Sprintf("%s: %s - %s", name, status, upd.Error)
	}
	return fmt.Sprintf("%s: %s", name, status)
}

func selectedPackages(selected map[string]bool) []config.Package {
	var pkgs []config.Package
	for _, pkg := range config.AllPackages() {
		key := pkg.Name
		if pkg.Required || selected[key] {
			pkgs = append(pkgs, pkg)
		}
	}
	sort.Slice(pkgs, func(i, j int) bool {
		if pkgs[i].Category == pkgs[j].Category {
			return strings.Compare(pkgs[i].Name, pkgs[j].Name) < 0
		}
		return strings.Compare(pkgs[i].Category, pkgs[j].Category) < 0
	})
	return pkgs
}

func timed(ctx context.Context, verbose bool, fn func(context.Context) (InstallStatus, string, error)) (InstallStatus, string, string, time.Duration) {
	start := time.Now()
	st, msg, err := fn(ctx)
	d := time.Since(start)
	if verbose {
		fmt.Printf("DEBUG: Operation took %s\n", d)
	}
	if err == nil {
		return st, msg, "", d
	}
	return StatusFailed, msg, err.Error(), d
}
