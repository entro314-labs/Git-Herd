package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	"githerd/internal/git"
	"githerd/internal/tui"
	"githerd/pkg/types"
)

// Manager handles bulk git operations with worker pools
type Manager struct {
	config    *types.Config
	logger    *slog.Logger
	scanner   *git.Scanner
	processor *git.Processor
}

// New creates a new Manager instance
func New(config *types.Config) *Manager {
	level := slog.LevelInfo
	if config.Verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return &Manager{
		config:    config,
		logger:    slog.New(handler),
		scanner:   git.NewScanner(config),
		processor: git.NewProcessor(config),
	}
}

// Execute runs the bulk git operation
func (m *Manager) Execute(ctx context.Context, rootPath string) error {
	// Use TUI if not in plain mode and not verbose (TUI doesn't work well with verbose logging)
	if !m.config.PlainMode && !m.config.Verbose {
		fmt.Printf("🎨 Starting GitHerd TUI...\n") // Debug
		model := tui.NewModel(m.config, rootPath)
		p := tea.NewProgram(model)
		
		if _, err := p.Run(); err != nil {
			fmt.Printf("TUI failed: %v\n", err) // Debug
			// Fallback to plain mode if TUI fails
			return m.executeInPlainMode(ctx, rootPath)
		}
		return nil
	}
	
	fmt.Printf("Using plain mode (plain: %v, verbose: %v)\n", m.config.PlainMode, m.config.Verbose) // Debug
	return m.executeInPlainMode(ctx, rootPath)
}

// executeInPlainMode runs the operation with plain text output
func (m *Manager) executeInPlainMode(ctx context.Context, rootPath string) error {
	m.logger.InfoContext(ctx, "Starting bulk git operation",
		"operation", m.config.Operation,
		"path", rootPath,
		"workers", m.config.Workers)

	// Find all git repositories
	repos, err := m.scanner.FindRepos(ctx, rootPath)
	if err != nil {
		return fmt.Errorf("failed to find repositories: %w", err)
	}

	if len(repos) == 0 {
		m.logger.InfoContext(ctx, "No git repositories found")
		return nil
	}

	m.logger.InfoContext(ctx, "Found repositories", "count", len(repos))

	// Process repositories concurrently
	return m.processReposConcurrently(ctx, repos)
}

// processReposConcurrently processes repositories using worker pools
func (m *Manager) processReposConcurrently(ctx context.Context, repos []types.GitRepo) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(m.config.Workers)

	resultChan := make(chan types.GitRepo, len(repos))

	// Start workers
	for _, repo := range repos {
		repo := repo // capture loop variable
		g.Go(func() error {
			processedRepo := m.processor.ProcessRepo(ctx, repo)
			select {
			case resultChan <- processedRepo:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Start result collector
	go func() {
		defer close(resultChan)
		_ = g.Wait() // Wait for all workers to complete
	}()

	// Collect and display results
	return m.displayResults(ctx, resultChan, len(repos))
}

// displayResults shows the results of the operations
func (m *Manager) displayResults(ctx context.Context, resultChan <-chan types.GitRepo, total int) error {
	var successful, failed, skipped int
	var allResults []types.GitRepo

	fmt.Printf("\n📊 Processing Results:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for result := range resultChan {
		allResults = append(allResults, result)
		
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "skipped") {
				skipped++
			} else {
				failed++
			}
			if m.config.FullSummary {
				fmt.Printf("❌ %s (%s): %v\n", result.Name, result.Path, result.Error)
			}
		} else {
			successful++
			status := "✅"
			if m.config.DryRun {
				status = "🔍"
			}
			if m.config.FullSummary {
				fmt.Printf("%s %s (%s) [%s@%s] - %v\n",
					status, result.Name, result.Path, result.Branch, result.Remote, result.Duration.Truncate(time.Millisecond))
			}
		}
	}

	// Show condensed view if not full summary
	if !m.config.FullSummary {
		// Show only first few and last few results
		displayCount := 5
		if len(allResults) <= displayCount*2 {
			displayCount = len(allResults) / 2
		}
		
		for i, result := range allResults[:displayCount] {
			m.displaySingleResult(result, i == 0)
		}
		
		if len(allResults) > displayCount*2 {
			fmt.Printf("... (%d more repositories) ...\n", len(allResults)-displayCount*2)
		}
		
		if len(allResults) > displayCount {
			start := len(allResults) - displayCount
			if len(allResults) <= displayCount*2 {
				start = displayCount
			}
			for _, result := range allResults[start:] {
				m.displaySingleResult(result, false)
			}
		}
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📈 Summary: %d successful, %d failed, %d skipped, %d total\n", successful, failed, skipped, total)

	// Save report to file if requested
	if m.config.SaveReport != "" {
		if err := m.saveReport(allResults, successful, failed, skipped); err != nil {
			m.logger.InfoContext(ctx, "Failed to save report", "error", err)
		} else {
			fmt.Printf("📄 Detailed report saved to: %s\n", m.config.SaveReport)
		}
	}

	if !m.config.FullSummary && len(allResults) > 10 {
		fmt.Printf("💡 Use --full-summary flag to see all %d repositories\n", len(allResults))
	}

	if failed > 0 {
		return fmt.Errorf("%d repositories failed", failed)
	}

	return nil
}

// displaySingleResult displays a single repository result
func (m *Manager) displaySingleResult(result types.GitRepo, isFirst bool) {
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "skipped") {
			fmt.Printf("⊝ %s (%s): %v\n", result.Name, result.Path, result.Error)
		} else {
			fmt.Printf("❌ %s (%s): %v\n", result.Name, result.Path, result.Error)
		}
	} else {
		status := "✅"
		if m.config.DryRun {
			status = "🔍"
		}
		fmt.Printf("%s %s (%s) [%s@%s] - %v\n",
			status, result.Name, result.Path, result.Branch, result.Remote, result.Duration.Truncate(time.Millisecond))
	}
}

// saveReport saves a detailed report to a file
func (m *Manager) saveReport(results []types.GitRepo, successful, failed, skipped int) error {
	file, err := os.Create(m.config.SaveReport)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Write header
	if _, err := fmt.Fprintf(file, "GitHerd Report - %s\n", time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("failed to write report header: %w", err)
	}
	if _, err := fmt.Fprintf(file, "Operation: %s\n", m.config.Operation); err != nil {
		return fmt.Errorf("failed to write operation: %w", err)
	}
	if _, err := fmt.Fprintf(file, "Workers: %d\n", m.config.Workers); err != nil {
		return fmt.Errorf("failed to write workers: %w", err)
	}
	if _, err := fmt.Fprintf(file, "Total Repositories: %d\n", len(results)); err != nil {
		return fmt.Errorf("failed to write total repositories: %w", err)
	}
	if _, err := fmt.Fprintf(file, "Successful: %d, Failed: %d, Skipped: %d\n\n", successful, failed, skipped); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	if _, err := fmt.Fprintf(file, "Repository Details:\n"); err != nil {
		return fmt.Errorf("failed to write details header: %w", err)
	}
	if _, err := fmt.Fprintf(file, "==================\n\n"); err != nil {
		return fmt.Errorf("failed to write details separator: %w", err)
	}

	for _, result := range results {
		if _, err := fmt.Fprintf(file, "Repository: %s\n", result.Name); err != nil {
			return fmt.Errorf("failed to write repository name: %w", err)
		}
		if _, err := fmt.Fprintf(file, "Path: %s\n", result.Path); err != nil {
			return fmt.Errorf("failed to write repository path: %w", err)
		}
		
		if result.Branch != "" {
			if _, err := fmt.Fprintf(file, "Branch: %s\n", result.Branch); err != nil {
				return fmt.Errorf("failed to write branch: %w", err)
			}
		}
		if result.Remote != "" {
			if _, err := fmt.Fprintf(file, "Remote: %s\n", result.Remote); err != nil {
				return fmt.Errorf("failed to write remote: %w", err)
			}
		}
		
		if _, err := fmt.Fprintf(file, "Duration: %v\n", result.Duration.Truncate(time.Millisecond)); err != nil {
			return fmt.Errorf("failed to write duration: %w", err)
		}
		
		if result.Error != nil {
			if _, err := fmt.Fprintf(file, "Status: FAILED - %v\n", result.Error); err != nil {
				return fmt.Errorf("failed to write failed status: %w", err)
			}
		} else if m.config.DryRun {
			if _, err := fmt.Fprintf(file, "Status: DRY RUN - Would have succeeded\n"); err != nil {
				return fmt.Errorf("failed to write dry run status: %w", err)
			}
		} else {
			if _, err := fmt.Fprintf(file, "Status: SUCCESS\n"); err != nil {
				return fmt.Errorf("failed to write success status: %w", err)
			}
		}
		
		if _, err := fmt.Fprintf(file, "\n"); err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}
	}

	return nil
}