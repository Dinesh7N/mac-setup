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

func IsBrewInstalled(ctx context.Context, verbose bool) bool {
	brewCmd := GetBrewExecutable()
	_, err := utils.Run(ctx, verbose, 5*time.Second, brewCmd, "--version")
	return err == nil
}

func InstallBrew(ctx context.Context, verbose bool) error {
	dir := os.TempDir()
	script := filepath.Join(dir, "macsetup-homebrew-install.sh")

	if err := utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, verbose, 0, "curl", "-fsSL", "-o", script, constants.HomebrewInstallScriptURL)
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

func BrewUpdate(ctx context.Context, verbose bool) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "update")
		return err
	})
}

func BrewUpgrade(ctx context.Context, verbose bool) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		_, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "upgrade")
		return err
	})
}

func AddTap(ctx context.Context, verbose bool, tap string) error {
	if tap == "" {
		return nil
	}
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 300 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "tap", tap)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func IsTapInstalled(ctx context.Context, verbose bool, tap string) (bool, error) {
	if tap == "" {
		return true, nil
	}
	res, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "tap")
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

func InstallFormula(ctx context.Context, verbose bool, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "install", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func LinkFormula(ctx context.Context, verbose bool, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	res, err := utils.Run(ctx, verbose, 10*time.Second, GetBrewExecutable(), "link", "--overwrite", name)
	if err != nil {
		// Check if it's already linked
		if strings.Contains(res.Stderr, "already linked") || strings.Contains(res.Stdout, "already linked") {
			return nil
		}
		if strings.TrimSpace(res.Stderr) != "" {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
		}
		return err
	}
	return nil
}

func ReinstallFormula(ctx context.Context, verbose bool, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "reinstall", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func InstallCask(ctx context.Context, verbose bool, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "install", "--cask", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func ReinstallCask(ctx context.Context, verbose bool, name string) error {
	brewMutex.Lock()
	defer brewMutex.Unlock()
	return utils.Retry(ctx, verbose, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "reinstall", "--cask", name)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}

func IsBrewPackageInstalled(ctx context.Context, verbose bool, pkg config.Package) (bool, error) {
	switch pkg.Type {
	case config.TypeFormula:
		_, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "list", "--formula", pkg.Name)
		return err == nil, nil
	case config.TypeCask:
		_, err := utils.Run(ctx, verbose, 0, GetBrewExecutable(), "list", "--cask", pkg.Name)
		return err == nil, nil
	default:
		return false, nil
	}
}

// IsCaskAppInstalled checks if a cask's app exists in /Applications
// This detects manually-installed apps that aren't managed by Homebrew
func IsCaskAppInstalled(caskName string) (bool, string) {
	// Map common cask names to their .app names
	caskToApp := map[string]string{
		"iterm2":    "iTerm.app",
		"raycast":   "Raycast.app",
		"rectangle": "Rectangle.app",
		"alacritty": "Alacritty.app",
		"ghostty":   "Ghostty.app",
		"docker":    "Docker.app",
		"slack":     "Slack.app",
		"zoom":      "zoom.us.app",
		"postman":   "Postman.app",
	}

	// Try to get app name from map, or use capitalized cask name + .app
	appName, ok := caskToApp[caskName]
	if !ok {
		// Default: capitalize first letter and add .app
		if len(caskName) > 0 {
			appName = strings.ToUpper(caskName[0:1]) + caskName[1:] + ".app"
		}
	}

	appPath := filepath.Join("/Applications", appName)
	if _, err := os.Stat(appPath); err == nil {
		return true, appPath
	}

	return false, ""
}

// IsFormulaLinked checks if a formula is properly linked in /opt/homebrew/bin
func IsFormulaLinked(ctx context.Context, verbose bool, name string) bool {
	// Use brew info to check if the package has a linked_keg
	res, err := utils.Run(ctx, verbose, 5*time.Second, GetBrewExecutable(), "info", "--json=v2", name)
	if err != nil {
		return false
	}

	// Check if linked_keg is not null in the JSON output
	// linked_keg will be the version string if linked, or null if not
	return strings.Contains(res.Stdout, `"linked_keg":`) && !strings.Contains(res.Stdout, `"linked_keg": null`)
}
