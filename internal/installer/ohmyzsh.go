package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"macsetup/internal/constants"
	"macsetup/internal/utils"
)

var zshPlugins = map[string]string{
	"zsh-autosuggestions":     "https://github.com/zsh-users/zsh-autosuggestions.git",
	"zsh-autocomplete":        "https://github.com/marlonrichert/zsh-autocomplete.git",
	"zsh-syntax-highlighting": "https://github.com/zsh-users/zsh-syntax-highlighting.git",
	"zsh-completions":         "https://github.com/zsh-users/zsh-completions.git",
}

func IsOhMyZshInstalled() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(filepath.Join(home, ".oh-my-zsh"))
	return err == nil, nil
}

func InstallOhMyZsh(ctx context.Context) error {
	installed, err := IsOhMyZshInstalled()
	if err != nil {
		return err
	}
	if installed {
		return nil
	}
	script, err := os.CreateTemp("", "macsetup-ohmyzsh-install.*.sh")
	if err != nil {
		return err
	}
	scriptPath := script.Name()
	_ = script.Close()
	defer func() {
		_ = os.Remove(scriptPath)
	}()

	if err := utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, "curl", "-fsSL", "-o", scriptPath, constants.OhMyZshInstallURL)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	_, err = utils.Run(ctx, 0, "sh", scriptPath, "", "--unattended")
	return err
}

func InstallZshPlugins(ctx context.Context) (InstallStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return StatusFailed, err
	}
	// Use custom plugins directory as requested
	pluginsDir := filepath.Join(home, ".oh-my-zsh", "custom", "plugins")

	allPresent := true
	for name := range zshPlugins {
		dest := filepath.Join(pluginsDir, name)
		if !utils.Exists(dest) {
			allPresent = false
			break
		}
	}
	if allPresent {
		return StatusSkipped, nil
	}

	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return StatusFailed, err
	}

	for name, url := range zshPlugins {
		dest := filepath.Join(pluginsDir, name)
		if utils.Exists(dest) {
			continue
		}
		if err := GitClone(ctx, url, dest); err != nil {
			return StatusFailed, fmt.Errorf("failed to clone %s: %w", name, err)
		}
	}
	return StatusInstalled, nil
}
