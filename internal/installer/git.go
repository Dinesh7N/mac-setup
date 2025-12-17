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

func CloneNeovimConfig(ctx context.Context) (InstallStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return StatusFailed, err
	}
	dest := filepath.Join(home, ".config", "nvim")
	if utils.Exists(dest) {
		return StatusSkipped, nil
	}
	if err := GitClone(ctx, constants.KickstartNvimURL, dest); err != nil {
		return StatusFailed, err
	}
	return StatusInstalled, nil
}

func CloneTPM(ctx context.Context) (InstallStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return StatusFailed, err
	}
	dest := filepath.Join(home, ".config", "tmux", "plugins", "tpm")
	if utils.Exists(dest) {
		return StatusSkipped, nil
	}
	if err := GitClone(ctx, constants.TpmURL, dest); err != nil {
		return StatusFailed, err
	}
	return StatusInstalled, nil
}

func GitClone(ctx context.Context, url, dest string) error {
	return utils.Retry(ctx, utils.RetryOptions{Attempts: 3, BaseDelay: 500 * time.Millisecond}, func(ctx context.Context) error {
		res, err := utils.Run(ctx, 0, "git", "clone", url, dest)
		if err != nil {
			if strings.TrimSpace(res.Stderr) != "" {
				return fmt.Errorf("%w: %s", err, strings.TrimSpace(res.Stderr))
			}
			return err
		}
		return nil
	})
}
