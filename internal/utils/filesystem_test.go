package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ExpandHome("~/x")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "x")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestWriteWithBackup(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(dest, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := WriteWithBackup(dest, []byte("new"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if backup == "" {
		t.Fatalf("expected backup path")
	}
	if !Exists(backup) {
		t.Fatalf("backup missing: %s", backup)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("dest content: got %q want %q", string(got), "new")
	}
}

func TestWriteWithBackupUnique(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(dest, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	b1, err := WriteWithBackup(dest, []byte("new1"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := WriteWithBackup(dest, []byte("new2"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if b1 == "" || b2 == "" || b1 == b2 {
		t.Fatalf("expected distinct backups, got %q and %q", b1, b2)
	}
}

func TestSymlinkIfMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SymlinkIfMissing(link, target); err != nil {
		t.Fatal(err)
	}
	if err := SymlinkIfMissing(link, target); err != nil {
		t.Fatal(err)
	}
}
