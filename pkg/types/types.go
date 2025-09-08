package types

import (
	"time"
)

// OperationType defines the type of git operation to perform
type OperationType string

const (
	OperationFetch OperationType = "fetch"
	OperationPull  OperationType = "pull"
)

// GitRepo represents a git repository with its path and status
type GitRepo struct {
	Path     string
	Name     string
	HasGit   bool
	Clean    bool
	Branch   string
	Remote   string
	Error    error
	Duration time.Duration
}

// Config holds application configuration
type Config struct {
	Workers     int
	Operation   OperationType
	DryRun      bool
	Recursive   bool
	SkipDirty   bool
	Verbose     bool
	Timeout     time.Duration
	ExcludeDirs []string
	PlainMode   bool  // Disable TUI for plain text output
	FullSummary bool  // Show full summary of all repositories
	SaveReport  string // File path to save detailed report
}

// GitRepoResult represents the result of processing a git repository
type GitRepoResult struct {
	Repo      GitRepo
	Success   bool
	Skipped   bool
	StartTime time.Time
	EndTime   time.Time
}

// ProcessingStats holds statistics about the processing session
type ProcessingStats struct {
	Total      int
	Successful int
	Failed     int
	Skipped    int
	StartTime  time.Time
	EndTime    time.Time
}

// Summary returns a formatted summary of the stats
func (s *ProcessingStats) Summary() string {
	return ""
}