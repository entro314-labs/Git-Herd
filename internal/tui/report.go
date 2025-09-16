package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/entro314-labs/Git-Herd/pkg/types"
)

// saveReport saves a detailed report to a file
func saveReport(config *types.Config, results []types.GitRepo, successful, failed, skipped int) error {
	file, err := os.Create(config.SaveReport)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Write header
	_, _ = fmt.Fprintf(file, "GitHerd Report - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	_, _ = fmt.Fprintf(file, "Operation: %s\n", config.Operation)
	_, _ = fmt.Fprintf(file, "Workers: %d\n", config.Workers)
	_, _ = fmt.Fprintf(file, "Total Repositories: %d\n", len(results))
	_, _ = fmt.Fprintf(file, "Successful: %d, Failed: %d, Skipped: %d\n\n", successful, failed, skipped)

	_, _ = fmt.Fprintf(file, "Repository Details:\n")
	_, _ = fmt.Fprintf(file, "==================\n\n")

	for _, result := range results {
		_, _ = fmt.Fprintf(file, "Repository: %s\n", result.Name)
		_, _ = fmt.Fprintf(file, "Path: %s\n", result.Path)
		
		if result.Branch != "" {
			_, _ = fmt.Fprintf(file, "Branch: %s\n", result.Branch)
		}
		if result.Remote != "" {
			_, _ = fmt.Fprintf(file, "Remote: %s\n", result.Remote)
		}
		
		_, _ = fmt.Fprintf(file, "Duration: %v\n", result.Duration.Truncate(time.Millisecond))
		
		if result.Error != nil {
			_, _ = fmt.Fprintf(file, "Status: FAILED - %v\n", result.Error)
		} else if config.DryRun {
			_, _ = fmt.Fprintf(file, "Status: DRY RUN - Would have succeeded\n")
		} else {
			_, _ = fmt.Fprintf(file, "Status: SUCCESS\n")
		}
		
		_, _ = fmt.Fprintf(file, "\n")
	}

	return nil
}