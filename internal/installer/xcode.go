package installer

import (
	"context"
	"time"

	"macsetup/internal/utils"
)

func IsXcodeInstalled(ctx context.Context) bool {
	_, err := utils.Run(ctx, false, 10*time.Second, "xcode-select", "-p")
	return err == nil
}

func TriggerXcodeInstall(ctx context.Context) error {
	_, _ = utils.Run(ctx, false, 0, "xcode-select", "--install")
	return nil
}

func WaitForXcode(ctx context.Context, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		if IsXcodeInstalled(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
