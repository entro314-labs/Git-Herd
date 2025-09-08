package git

import (
	"context"
	"fmt"
	"time"

	gogit "github.com/go-git/go-git/v5"

	"githerd/pkg/types"
)

// Processor handles git operations on repositories
type Processor struct {
	config *types.Config
}

// NewProcessor creates a new git operations processor
func NewProcessor(config *types.Config) *Processor {
	return &Processor{
		config: config,
	}
}

// AnalyzeRepo analyzes a git repository to determine its status
func (p *Processor) AnalyzeRepo(repo *types.GitRepo) {
	start := time.Now()
	defer func() {
		repo.Duration = time.Since(start)
	}()

	gitRepo, err := gogit.PlainOpen(repo.Path)
	if err != nil {
		repo.Error = fmt.Errorf("failed to open repository: %w", err)
		return
	}

	// Get current branch
	head, err := gitRepo.Head()
	if err != nil {
		repo.Error = fmt.Errorf("failed to get HEAD: %w", err)
		return
	}

	if head.Name().IsBranch() {
		repo.Branch = head.Name().Short()
	} else {
		repo.Branch = "detached"
	}

	// Check working tree status
	worktree, err := gitRepo.Worktree()
	if err != nil {
		repo.Error = fmt.Errorf("failed to get worktree: %w", err)
		return
	}

	status, err := worktree.Status()
	if err != nil {
		repo.Error = fmt.Errorf("failed to get status: %w", err)
		return
	}

	repo.Clean = status.IsClean()

	// Get remote information
	remotes, err := gitRepo.Remotes()
	if err == nil && len(remotes) > 0 {
		repo.Remote = remotes[0].Config().Name
	}
}

// ProcessRepo performs the git operation on a single repository
func (p *Processor) ProcessRepo(ctx context.Context, repo types.GitRepo) types.GitRepo {
	start := time.Now()
	defer func() {
		repo.Duration = time.Since(start)
	}()

	// Analyze repo first (moved from scanning phase for better performance)
	p.AnalyzeRepo(&repo)

	if repo.Error != nil {
		return repo
	}

	// Skip dirty repos if configured
	if p.config.SkipDirty && !repo.Clean {
		repo.Error = fmt.Errorf("repository has uncommitted changes (skipped)")
		return repo
	}

	if p.config.DryRun {
		return repo
	}

	gitRepo, err := gogit.PlainOpen(repo.Path)
	if err != nil {
		repo.Error = fmt.Errorf("failed to open repository: %w", err)
		return repo
	}

	switch p.config.Operation {
	case types.OperationFetch:
		err = p.fetchRepo(ctx, gitRepo)
	case types.OperationPull:
		err = p.pullRepo(ctx, gitRepo)
	}

	if err != nil {
		repo.Error = err
	}

	return repo
}

// fetchRepo performs git fetch on a repository
func (p *Processor) fetchRepo(ctx context.Context, repo *gogit.Repository) error {
	err := repo.FetchContext(ctx, &gogit.FetchOptions{
		RemoteName: "origin",
		Progress:   nil, // We could add progress reporting here
	})

	if err != nil && err != gogit.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch failed: %w", err)
	}

	return nil
}

// pullRepo performs git pull on a repository
func (p *Processor) pullRepo(ctx context.Context, repo *gogit.Repository) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.PullContext(ctx, &gogit.PullOptions{
		RemoteName: "origin",
		Progress:   nil,
	})

	if err != nil && err != gogit.NoErrAlreadyUpToDate {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}