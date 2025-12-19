package utils

import (
	"context"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	res, err := Run(context.Background(), false, 1*time.Second, "/bin/sh", "-c", "printf hello")
	if err != nil {
		t.Fatal(err)
	}
	if res.Stdout != "hello" {
		t.Fatalf("stdout: got %q want %q", res.Stdout, "hello")
	}
}

func TestRunTimeout(t *testing.T) {
	_, err := Run(context.Background(), false, 10*time.Millisecond, "/bin/sh", "-c", "sleep 1")
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}
