package installer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"macsetup/configs"
	"macsetup/internal/utils"
)

type zshrcTemplateData struct {
	Timestamp string
}

func WriteDotfiles() (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	type fileSpec struct {
		src       string
		dest      string
		mode      os.FileMode
		isTmpl    bool
		tmplData  any
		ensureDir bool
	}

	files := []fileSpec{
		{
			src:       "zshrc.tmpl",
			dest:      filepath.Join(home, ".zshrc"),
			mode:      0o644,
			isTmpl:    true,
			tmplData:  zshrcTemplateData{Timestamp: time.Now().Format(time.RFC3339)},
			ensureDir: false,
		},
		{
			src:       "starship.toml",
			dest:      filepath.Join(home, ".config", "starship", "starship.toml"),
			mode:      0o644,
			ensureDir: true,
		},
		{
			src:       "alacritty.toml",
			dest:      filepath.Join(home, ".config", "alacritty", "alacritty.toml"),
			mode:      0o644,
			ensureDir: true,
		},
		{
			src:       "ghostty.conf",
			dest:      filepath.Join(home, ".config", "ghostty", "config"),
			mode:      0o644,
			ensureDir: true,
		},
		{
			src:       "tmux.conf",
			dest:      filepath.Join(home, ".config", "tmux", "tmux.conf"),
			mode:      0o644,
			ensureDir: true,
		},
		{
			src:       "zellij.kdl",
			dest:      filepath.Join(home, ".config", "zellij", "config.kdl"),
			mode:      0o644,
			ensureDir: true,
		},
	}

	backups := 0
	for _, f := range files {
		if f.ensureDir {
			if err := os.MkdirAll(filepath.Dir(f.dest), 0o755); err != nil {
				return backups, err
			}
		}

		content, err := configs.FS.ReadFile(f.src)
		if err != nil {
			return backups, err
		}

		if f.isTmpl {
			tmpl, err := template.New(f.src).Parse(string(content))
			if err != nil {
				return backups, err
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, f.tmplData); err != nil {
				return backups, err
			}
			content = buf.Bytes()
		}

		backup, err := utils.WriteWithBackup(f.dest, content, f.mode)
		if err != nil {
			return backups, err
		}
		if backup != "" {
			backups++
		}
	}

	tmuxConf := filepath.Join(home, ".tmux.conf")
	tmuxTarget := filepath.Join(home, ".config", "tmux", "tmux.conf")
	if err := utils.SymlinkIfMissing(tmuxConf, tmuxTarget); err != nil {
		return backups, fmt.Errorf("failed to create tmux symlink: %w", err)
	}

	return backups, nil
}
