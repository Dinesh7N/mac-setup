package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySucceedsAfterFailures(t *testing.T) {
	var calls int
	err := Retry(context.Background(), false, RetryOptions{Attempts: 3, BaseDelay: 1 * time.Millisecond, MaxDelay: 2 * time.Millisecond}, func(ctx context.Context) error {
		calls++
		if calls < 2 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls: got %d want %d", calls, 2)
	}
}

func TestRetryReturnsLastError(t *testing.T) {
	var calls int
	want := errors.New("nope")
	err := Retry(context.Background(), false, RetryOptions{Attempts: 3, BaseDelay: 1 * time.Millisecond, MaxDelay: 2 * time.Millisecond}, func(ctx context.Context) error {
		calls++
		return want
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected errors.Is(err, want)")
	}
	if calls != 3 {
		t.Fatalf("calls: got %d want %d", calls, 3)
	}
}

func TestRetryHonorsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Retry(ctx, false, RetryOptions{Attempts: 3, BaseDelay: 1 * time.Millisecond, MaxDelay: 2 * time.Millisecond}, func(ctx context.Context) error {
		return errors.New("temporary")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
