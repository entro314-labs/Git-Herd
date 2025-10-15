package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/entro314-labs/git-herd/internal/git"
	"github.com/entro314-labs/git-herd/pkg/types"
)

type Model struct {
	config    *types.Config
	rootPath  string
	ctx       context.Context
	cancel    context.CancelFunc
	scanner   *git.Scanner
	processor *git.Processor

	// UI state
	phase     string
	spinner   spinner.Model
	progress  progress.Model
	repos     []types.GitRepo
	processed int
	results   []types.GitRepo

	// Status
	scanning   bool
	processing bool
	done       bool
	err        error
}

type reposFoundMsg []types.GitRepo
type repoProcessedMsg types.GitRepo
type processingDoneMsg struct {
	err error
}

func NewModel(config *types.Config, rootPath string) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	p := progress.New(progress.WithDefaultGradient())

	return &Model{
		config:    config,
		rootPath:  rootPath,
		ctx:       ctx,
		cancel:    cancel,
		scanner:   git.NewScanner(config),
		processor: git.NewProcessor(config),
		phase:     "initializing",
		spinner:   s,
		progress:  p,
		scanning:  true,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.scanRepos(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel()
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case reposFoundMsg:
		m.repos = []types.GitRepo(msg)
		m.scanning = false
		m.processing = true
		m.phase = "processing"

		if len(m.repos) == 0 {
			m.done = true
			m.phase = "complete"
			return m, tea.Quit
		}

		return m, m.processRepos()

	case repoProcessedMsg:
		m.results = append(m.results, types.GitRepo(msg))
		m.processed++

		if m.processed >= len(m.repos) {
			m.processing = false
			m.done = true
			m.phase = "complete"
			return m, tea.Sequence(
				tea.Printf("\n"),
				tea.Quit,
			)
		}

		// Process next repo
		return m, m.processNextRepo()

	case processingDoneMsg:
		m.processing = false
		m.done = true
		m.phase = "complete"
		m.err = msg.err
		return m, tea.Sequence(
			tea.Printf("\n"),
			tea.Quit,
		)
	}

	return m, nil
}

func (m *Model) scanRepos() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		repos, err := m.scanner.FindRepos(m.ctx, m.rootPath)
		if err != nil {
			return processingDoneMsg{err: err}
		}
		return reposFoundMsg(repos)
	})
}

func (m *Model) processRepos() tea.Cmd {
	return func() tea.Msg {
		// Process first repo
		if len(m.repos) > 0 {
			processed := m.processor.ProcessRepo(m.ctx, m.repos[0])
			return repoProcessedMsg(processed)
		}

		return processingDoneMsg{err: nil}
	}
}

func (m *Model) processNextRepo() tea.Cmd {
	return func() tea.Msg {
		if m.processed < len(m.repos) {
			processed := m.processor.ProcessRepo(m.ctx, m.repos[m.processed])
			return repoProcessedMsg(processed)
		}
		return processingDoneMsg{err: nil}
	}
}
