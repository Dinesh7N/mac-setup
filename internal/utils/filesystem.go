package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ExpandHome(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	return "", fmt.Errorf("unsupported tilde path: %q", path)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func EnsureDir(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func WriteWithBackup(dest string, content []byte, mode os.FileMode) (string, error) {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(dest)+".tmp.*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	if _, err := tmp.Write(content); err != nil {
		cleanup()
		return "", err
	}
	if err := tmp.Chmod(mode); err != nil {
		cleanup()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", err
	}

	var backup string
	if Exists(dest) {
		backup = backupPath(dest)
		if err := os.Rename(dest, backup); err != nil {
			cleanup()
			return "", fmt.Errorf("failed to backup %s: %w", dest, err)
		}
	}

	if err := os.Rename(tmpName, dest); err != nil {
		if backup != "" {
			_ = os.Rename(backup, dest)
		}
		cleanup()
		return backup, err
	}
	return backup, nil
}

func SymlinkIfMissing(linkPath, target string) error {
	if _, err := os.Lstat(linkPath); err == nil {
		return nil
	}
	return os.Symlink(target, linkPath)
}

func backupPath(dest string) string {
	timestamp := time.Now().Format("20060102_150405.000")
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return fmt.Sprintf("%s.bak.%s.%s", dest, timestamp, hex.EncodeToString(buf[:]))
}
