package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"macsetup/internal/config"
	"macsetup/internal/constants"
	"macsetup/internal/utils"
)

// brewMutex ensures only one brew command runs at a time to avoid Homebrew's file locking issues
var brewMutex sync.Mutex

func GetBrewExecutable() string {
	if path, err := exec.LookPath("brew"); err == nil {
		return path
	}
	return "/opt/homebrew/bin/brew"
}

func IsBrewInstalled(ctx context.Context) bool {
	brewCmd := GetBrewExecutable()
	_, err := utils.Run(ctx, 5*time.Second, brewCmd, "--version")
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
	defer func() {
		_ = os.Remove(script)
	}()

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
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, 0, GetBrewExecutable(), "update")
		return err
	})
}

func BrewUpgrade(ctx context.Context) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, 0, GetBrewExecutable(), "upgrade")
		return err
	})
}

func AddTap(ctx context.Context, tap string) error {
	if tap == "" {
		return nil
	}
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, GetBrewExecutable(), "tap", tap)
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
	res, err := utils.Run(ctx, 0, GetBrewExecutable(), "tap")
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
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, GetBrewExecutable(), "install", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func ReinstallFormula(ctx context.Context, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, GetBrewExecutable(), "reinstall", name)
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
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, GetBrewExecutable(), "install", "--cask", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func ReinstallCask(ctx context.Context, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, GetBrewExecutable(), "reinstall", "--cask", name)
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
		_, err := utils.Run(ctx, 0, GetBrewExecutable(), "list", "--formula", pkg.Name)
		if err == nil {
			return true, nil
		}
		return false, nil
	case config.TypeCask:
		_, err := utils.Run(ctx, 0, GetBrewExecutable(), "list", "--cask", pkg.Name)
		if err == nil {
			return true, nil
		}
		return false, nil
	default:
		return false, nil
	}
}
