package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/entro314-labs/Git-Herd/pkg/types"
)

// Scanner handles discovering git repositories in a directory tree
type Scanner struct {
	config *types.Config
}

// NewScanner creates a new git repository scanner
func NewScanner(config *types.Config) *Scanner {
	return &Scanner{
		config: config,
	}
}

// FindRepos discovers all git repositories in the given directory
func (s *Scanner) FindRepos(ctx context.Context, rootPath string) ([]types.GitRepo, error) {
	var repos []types.GitRepo
	var mu sync.Mutex
	var foundCount int

	// Print initial scanning message
	if s.config.PlainMode || s.config.Verbose {
		fmt.Printf("🔍 Scanning for Git repositories in %s...\n", rootPath)
	}

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip if not a directory
		if !d.IsDir() {
			return nil
		}

		// Check if we should exclude this directory
		for _, exclude := range s.config.ExcludeDirs {
			if strings.Contains(path, exclude) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if this is a git repository
		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			repo := types.GitRepo{
				Path:   path,
				Name:   filepath.Base(path),
				HasGit: true,
			}

			// Don't analyze repo here - defer to processing phase for better performance
			mu.Lock()
			repos = append(repos, repo)
			foundCount++
			
			// Show progress every 10 repositories found (only in plain/verbose mode)
			if (s.config.PlainMode || s.config.Verbose) && foundCount%10 == 0 {
				fmt.Printf("   Found %d repositories so far...\n", foundCount)
			}
			mu.Unlock()

			// Skip subdirectories if not recursive
			if !s.config.Recursive {
				return filepath.SkipDir
			}
		}

		return nil
	})

	// Print final count
	if s.config.PlainMode || s.config.Verbose {
		fmt.Printf("✅ Scan complete: found %d Git repositories\n", len(repos))
	}

	return repos, err
}