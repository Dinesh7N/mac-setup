package utils

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

type RetryOptions struct {
	Attempts  int
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func Retry(ctx context.Context, verbose bool, opts RetryOptions, fn func(context.Context) error) error {
	attempts := opts.Attempts
	if attempts <= 0 {
		attempts = 1
	}
	delay := opts.BaseDelay
	if delay <= 0 {
		delay = 250 * time.Millisecond
	}
	maxDelay := opts.MaxDelay
	if maxDelay <= 0 {
		maxDelay = 5 * time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if verbose && attempt > 1 {
			fmt.Printf("Retrying (attempt %d/%d)...\n", attempt, attempts)
		}
		if err := fn(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt == attempts {
			break
		}

		sleep := delay + jitter(delay/4)
		if sleep > maxDelay {
			sleep = maxDelay
		}

		t := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
	return lastErr
}

func jitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0
	}
	n := binary.LittleEndian.Uint64(buf[:])
	return time.Duration(n % uint64(max))
}
