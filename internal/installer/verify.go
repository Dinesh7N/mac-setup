package installer

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"macsetup/internal/utils"
)

type VerifyResult struct {
	Name  string
	Error string
}

func VerifyCriticalTools(ctx context.Context) []VerifyResult {
	checks := []struct {
		name string
		cmd  []string
	}{
		{name: "brew", cmd: []string{"brew", "--version"}},
		{name: "git", cmd: []string{"git", "--version"}},
		{name: "nvim", cmd: []string{"nvim", "--version"}},
		{name: "tmux", cmd: []string{"tmux", "-V"}},
		{name: "mise", cmd: []string{"mise", "--version"}},
		{name: "starship", cmd: []string{"starship", "--version"}},
	}

	var results []VerifyResult
	for _, c := range checks {
		if _, err := exec.LookPath(c.cmd[0]); err != nil {
			results = append(results, VerifyResult{Name: c.name, Error: "not found in PATH"})
			continue
		}
		if _, err := utils.Run(ctx, 10*time.Second, c.cmd[0], c.cmd[1:]...); err != nil {
			results = append(results, VerifyResult{Name: c.name, Error: err.Error()})
			continue
		}
		results = append(results, VerifyResult{Name: c.name})
	}
	return results
}

func VerifySummary(results []VerifyResult) (ok int, failed int, msg string) {
	for _, r := range results {
		if r.Error != "" {
			failed++
		} else {
			ok++
		}
	}
	if failed == 0 {
		return ok, failed, ""
	}
	return ok, failed, fmt.Sprintf("%d failed verification checks", failed)
}
