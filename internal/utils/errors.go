package utils

import (
	"fmt"
	"strings"
)

type InstallErrorType string

const (
	ErrNetwork    InstallErrorType = "network"
	ErrPermission InstallErrorType = "permission"
	ErrDependency InstallErrorType = "dependency"
	ErrNotFound   InstallErrorType = "not_found"
	ErrTimeout    InstallErrorType = "timeout"
	ErrLock       InstallErrorType = "lock"
	ErrUnknown    InstallErrorType = "unknown"
)

type InstallError struct {
	Package string
	Type    InstallErrorType
	Message string
	Stderr  string
}

func (e *InstallError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Package, e.Message)
}

func ClassifyError(pkg string, err error, stderr string) *InstallError {
	ie := &InstallError{Package: pkg, Stderr: stderr}

	switch {
	case strings.Contains(stderr, "Could not resolve host"):
		ie.Type = ErrNetwork
		ie.Message = "Network error - check your internet connection"
	case strings.Contains(stderr, "Permission denied") || strings.Contains(stderr, "Operation not permitted"):
		ie.Type = ErrPermission
		ie.Message = "Permission denied - check macOS Seatbelt or try running with sudo"
	case strings.Contains(stderr, "No available formula"):
		ie.Type = ErrNotFound
		ie.Message = "Package not found in Homebrew"
	case strings.Contains(strings.ToLower(stderr), "dependency"):
		ie.Type = ErrDependency
		ie.Message = "Dependency conflict"
	case strings.Contains(stderr, "Another active Homebrew process is already in progress") || strings.Contains(stderr, "waiting for lock"):
		ie.Type = ErrLock
		ie.Message = "Homebrew is locked by another process (internal mutex should handle this)"
	default:
		ie.Type = ErrUnknown
		ie.Message = err.Error()
	}

	return ie
}
