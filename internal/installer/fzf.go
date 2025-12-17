package installer

import (
	"context"
	"os"
	"path/filepath"

	"macsetup/internal/utils"
)

func ConfigureFzf(ctx context.Context) (InstallStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return StatusFailed, err
	}
	if utils.Exists(filepath.Join(home, ".fzf.zsh")) {
		return StatusSkipped, nil
	}
	brewCmd := GetBrewExecutable()
	if _, err := utils.Run(ctx, 0, brewCmd, "list", "--formula", "fzf"); err != nil {
		return StatusSkipped, nil
	}
	res, err := utils.Run(ctx, 0, brewCmd, "--prefix")
	if err != nil {
		return StatusFailed, err
	}
	prefix := res.Stdout
	for len(prefix) > 0 && (prefix[len(prefix)-1] == '\n' || prefix[len(prefix)-1] == '\r') {
		prefix = prefix[:len(prefix)-1]
	}
	installScript := filepath.Join(prefix, "opt", "fzf", "install")
	_, err = utils.Run(ctx, 0, installScript, "--all", "--no-bash", "--no-fish")
	if err != nil {
		return StatusFailed, err
	}
	return StatusInstalled, nil
}
