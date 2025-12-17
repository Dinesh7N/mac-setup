package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"macsetup/internal/config"
	"macsetup/internal/constants"
	"macsetup/internal/utils"
)

func IsBrewInstalled(ctx context.Context) bool {
	_, err := exec.LookPath("brew")
	if err == nil {
		return true
	}
	_, err = utils.Run(ctx, 5*time.Second, "/opt/homebrew/bin/brew", "--version")
	return err == nil
}

func InstallBrew(ctx context.Context) error {
	dir := os.TempDir()
	script := filepath.Join(dir, "macsetup-homebrew-install.sh")

	if err := utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, 0, "curl", "-fsSL", "-o", script, constants.HomebrewInstallScriptURL)
		return err
	}); err != nil {
		return err
	}
	defer os.Remove(script)

	cmd := exec.CommandContext(ctx, "/bin/bash", script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	path := os.Getenv("PATH")
	if !strings.Contains(path, "/opt/homebrew/bin") {
		_ = os.Setenv("PATH", "/opt/homebrew/bin:"+path)
	}
	return nil
}

func BrewUpdate(ctx context.Context) error {
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, 0, "brew", "update")
		return err
	})
}

func BrewUpgrade(ctx context.Context) error {
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, 0, "brew", "upgrade")
		return err
	})
}

func AddTap(ctx context.Context, tap string) error {
	if tap == "" {
		return nil
	}
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, "brew", "tap", tap)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func IsTapInstalled(ctx context.Context, tap string) (bool, error) {
	if tap == "" {
		return true, nil
	}
	res, err := utils.Run(ctx, 0, "brew", "tap")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(res.Stdout, "\n") {
		if strings.TrimSpace(line) == tap {
			return true, nil
		}
	}
	return false, nil
}

func InstallFormula(ctx context.Context, name string) error {
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, "brew", "install", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func InstallCask(ctx context.Context, name string) error {
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, "brew", "install", "--cask", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func IsBrewPackageInstalled(ctx context.Context, pkg config.Package) (bool, error) {
	switch pkg.Type {
	case config.TypeFormula:
		_, err := utils.Run(ctx, 0, "brew", "list", "--formula", pkg.Name)
		if err == nil {
			return true, nil
		}
		return false, nil
	case config.TypeCask:
		_, err := utils.Run(ctx, 0, "brew", "list", "--cask", pkg.Name)
		if err == nil {
			return true, nil
		}
		return false, nil
	default:
		return false, nil
	}
}
