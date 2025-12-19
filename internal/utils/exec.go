package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CmdResult struct {
	Stdout string
	Stderr string
}

func Run(ctx context.Context, verbose bool, timeout time.Duration, name string, args ...string) (CmdResult, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	res := CmdResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if verbose {
		fmt.Printf("CMD: %s %s\n", name, strings.Join(args, " "))
		if stdout.Len() > 0 {
			fmt.Printf("STDOUT:\n%s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			fmt.Printf("STDERR:\n%s\n", stderr.String())
		}
	}
	if err == nil {
		return res, nil
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return res, fmt.Errorf("command timed out: %s %v", name, args)
	}
	return res, err
}
