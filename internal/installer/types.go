package installer

import (
	"time"

	"macsetup/internal/config"
)

type InstallStatus string

const (
	StatusPending   InstallStatus = "pending"
	StatusRunning   InstallStatus = "running"
	StatusInstalled InstallStatus = "installed"
	StatusSkipped   InstallStatus = "skipped"
	StatusFailed    InstallStatus = "failed"
)

type InstallResult struct {
	Package  config.Package
	Status   InstallStatus
	Message  string
	Error    string
	Duration time.Duration
}

type ProgressUpdate struct {
	Package config.Package
	Status  InstallStatus
	Message string
	Error   string
}

type Summary struct {
	Results []InstallResult
}

func (s Summary) FailedCount() int {
	n := 0
	for _, r := range s.Results {
		if r.Status == StatusFailed {
			n++
		}
	}
	return n
}
