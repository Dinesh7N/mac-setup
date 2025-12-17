package installer

import (
	"context"
	"strings"
	"time"

	"macsetup/internal/config"
	"macsetup/internal/utils"
)

// ScanInstalledPackages returns a map of package names that are currently installed.
func ScanInstalledPackages(ctx context.Context, packages []config.Package) (map[string]bool, error) {
	installed := make(map[string]bool)

	// 1. Scan Brew Formulas and Casks (Bulk)
	if IsBrewInstalled(ctx) {
		// Get all formulas
		if out, err := utils.Run(ctx, 10*time.Second, "brew", "list", "--formula", "-1"); err == nil {
			for _, line := range strings.Split(out.Stdout, "\n") {
				if name := strings.TrimSpace(line); name != "" {
					installed["formula:"+name] = true
				}
			}
		}

		// Get all casks
		if out, err := utils.Run(ctx, 10*time.Second, "brew", "list", "--cask", "-1"); err == nil {
			for _, line := range strings.Split(out.Stdout, "\n") {
				if name := strings.TrimSpace(line); name != "" {
					installed["cask:"+name] = true
				}
			}
		}
	}

	// 2. Check individual packages based on type
	for _, pkg := range packages {
		isInstalled := false

		switch pkg.Type {
		case config.TypeSystem:
			if pkg.Name == "Xcode CLI Tools" {
				isInstalled = IsXcodeInstalled(ctx)
			} else if pkg.Name == "Homebrew" {
				isInstalled = IsBrewInstalled(ctx)
			}
		case config.TypeFormula:
			// Checked via bulk list, but we need to match the name.
			// Sometimes brew list name differs slightly, but usually it's exact.
			// We check the map we populated.
			if installed["formula:"+pkg.Name] {
				isInstalled = true
			}
		case config.TypeCask:
			if installed["cask:"+pkg.Name] {
				isInstalled = true
			}
		}

		if isInstalled {
			installed[pkg.Name] = true
		}
	}

	return installed, nil
}
