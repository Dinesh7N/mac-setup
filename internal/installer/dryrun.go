package installer

import (
	"context"
	"fmt"

	"macsetup/internal/config"
)

func DryRunPlan(ctx context.Context, selected map[string]bool) []string {
	var lines []string

	if IsXcodeInstalled(ctx) {
		lines = append(lines, "Xcode CLI Tools: already installed (skip)")
	} else {
		lines = append(lines, "Xcode CLI Tools: would install (GUI prompt)")
	}

	brewInstalled := IsBrewInstalled(ctx)
	if brewInstalled {
		lines = append(lines, "Homebrew: already installed (skip install, would run brew update/upgrade)")
	} else {
		lines = append(lines, "Homebrew: would install, then run brew update/upgrade")
	}

	pkgs := selectedPackages(selected)
	taps, formulas, casks := splitBrewPackages(pkgs)
	if len(taps) > 0 {
		lines = append(lines, fmt.Sprintf("Homebrew taps: %d", len(taps)))
		for _, t := range taps {
			name := t.Tap
			if name == "" {
				name = t.Name
			}
			lines = append(lines, "  - "+formatDryRunTap(ctx, brewInstalled, name))
		}
	}
	if len(formulas) > 0 {
		lines = append(lines, fmt.Sprintf("Formulas (parallel): %d", len(formulas)))
		for _, f := range formulas {
			lines = append(lines, "  - "+formatDryRunBrewPkg(ctx, brewInstalled, f))
		}
	}
	if len(casks) > 0 {
		lines = append(lines, fmt.Sprintf("Casks (sequential): %d", len(casks)))
		for _, c := range casks {
			lines = append(lines, "  - "+formatDryRunBrewPkg(ctx, brewInstalled, c))
		}
	}

	lines = append(lines, "Post-install tasks:")
	lines = append(lines, "  - Create config directories")
	lines = append(lines, "  - Oh My Zsh + plugins")
	lines = append(lines, "  - Neovim config (skip if ~/.config/nvim exists)")
	lines = append(lines, "  - TPM (tmux plugins)")
	lines = append(lines, "  - Mise runtimes (node/python/go)")
	lines = append(lines, "  - Write dotfiles (with backups)")
	lines = append(lines, "  - Configure fzf (skip if already configured)")
	return lines
}

func formatDryRunTap(ctx context.Context, brewInstalled bool, tap string) string {
	if !brewInstalled {
		return fmt.Sprintf("%s: would tap (brew not installed yet)", tap)
	}
	ok, err := IsTapInstalled(ctx, tap)
	if err != nil {
		return fmt.Sprintf("%s: would tap (status unknown: %s)", tap, err.Error())
	}
	if ok {
		return fmt.Sprintf("%s: already tapped (skip)", tap)
	}
	return fmt.Sprintf("%s: would tap", tap)
}

func formatDryRunBrewPkg(ctx context.Context, brewInstalled bool, pkg config.Package) string {
	if !brewInstalled {
		return fmt.Sprintf("%s (%s): would install (brew not installed yet)", pkg.Name, pkg.Type)
	}
	ok, err := IsBrewPackageInstalled(ctx, pkg)
	if err != nil {
		return fmt.Sprintf("%s (%s): would install (status unknown: %s)", pkg.Name, pkg.Type, err.Error())
	}
	if ok {
		return fmt.Sprintf("%s (%s): already installed (skip)", pkg.Name, pkg.Type)
	}
	return fmt.Sprintf("%s (%s): would install", pkg.Name, pkg.Type)
}
