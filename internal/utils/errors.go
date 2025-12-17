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
	case strings.Contains(stderr, "Permission denied"):
		ie.Type = ErrPermission
		ie.Message = "Permission denied - try running with sudo"
	case strings.Contains(stderr, "No available formula"):
		ie.Type = ErrNotFound
		ie.Message = "Package not found in Homebrew"
	case strings.Contains(strings.ToLower(stderr), "dependency"):
		ie.Type = ErrDependency
		ie.Message = "Dependency conflict"
	default:
		ie.Type = ErrUnknown
		ie.Message = err.Error()
	}

	return ie
}
