package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateConfigDirectories() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		".config/starship",
		".config/alacritty",
		".config/ghostty",
		".config/tmux/plugins",
		".config/zellij",
		".config/nvim",
		".config/mise",
		".config/1Password/ssh",
		".config/op",
	}

	for _, dir := range dirs {
		path := filepath.Join(home, dir)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", path, err)
		}
	}
	return nil
}
