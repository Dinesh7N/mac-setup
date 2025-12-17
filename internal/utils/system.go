package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func PreflightChecks(ctx context.Context) error {
	if runtime.GOOS != "darwin" {
		return errors.New("macsetup only runs on macOS")
	}
	if runtime.GOARCH != "arm64" {
		return errors.New("macsetup only supports Apple Silicon (arm64)")
	}
	if err := ValidateDependencies(); err != nil {
		return err
	}
	if err := RequestSudo(); err != nil {
		return fmt.Errorf("sudo required: %w", err)
	}
	go KeepSudoAlive(ctx)
	return nil
}

func RequestSudo() error {
	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func KeepSudoAlive(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = exec.Command("sudo", "-n", "true").Run()
		}
	}
}

func ValidateDependencies() error {
	required := []string{"sudo", "curl", "git", "bash"}
	for _, name := range required {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("required dependency not found in PATH: %s", name)
		}
	}
	return nil
}
