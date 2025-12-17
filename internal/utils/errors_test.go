package utils

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyError(t *testing.T) {
	cases := []struct {
		name      string
		stderr    string
		wantType  InstallErrorType
		wantMsg   string
		wantInErr string
	}{
		{
			name:      "network",
			stderr:    "curl: (6) Could not resolve host: example.com",
			wantType:  ErrNetwork,
			wantMsg:   "Network error - check your internet connection",
			wantInErr: "[network]",
		},
		{
			name:      "permission",
			stderr:    "Permission denied",
			wantType:  ErrPermission,
			wantMsg:   "Permission denied - try running with sudo",
			wantInErr: "[permission]",
		},
		{
			name:      "not found",
			stderr:    "Error: No available formula with the name \"nope\"",
			wantType:  ErrNotFound,
			wantMsg:   "Package not found in Homebrew",
			wantInErr: "[not_found]",
		},
		{
			name:      "dependency",
			stderr:    "Error: dependency resolution failed",
			wantType:  ErrDependency,
			wantMsg:   "Dependency conflict",
			wantInErr: "[dependency]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ie := ClassifyError("pkg", errors.New("exit status 1"), tc.stderr)
			if ie.Type != tc.wantType {
				t.Fatalf("type: got %q want %q", ie.Type, tc.wantType)
			}
			if ie.Message != tc.wantMsg {
				t.Fatalf("message: got %q want %q", ie.Message, tc.wantMsg)
			}
			if got := ie.Error(); got == "" || !strings.Contains(got, tc.wantInErr) {
				t.Fatalf("error string: got %q, want to contain %q", got, tc.wantInErr)
			}
		})
	}
}
