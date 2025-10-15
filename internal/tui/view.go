package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#01FAC6")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#02BA84")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61DAFB"))

	summaryStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Margin(1, 0)

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)

func (m *Model) View() string {
	if m.done {
		return m.renderSummary()
	}

	var content strings.Builder

	// Title
	titleCaser := cases.Title(language.English)
	title := fmt.Sprintf("git-herd - %s Operation", titleCaser.String(string(m.config.Operation)))
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Current phase
	switch m.phase {
	case "initializing", "scanning":
		content.WriteString(fmt.Sprintf("%s Scanning for Git repositories in %s\n",
			m.spinner.View(),
			infoStyle.Render(m.rootPath)))

	case "processing":
		if len(m.repos) > 0 {
			percent := float64(m.processed) / float64(len(m.repos))
			content.WriteString(fmt.Sprintf("Processing repositories %s\n",
				statusStyle.Render(fmt.Sprintf("(%d/%d)", m.processed, len(m.repos)))))
			content.WriteString(m.progress.ViewAs(percent))
			content.WriteString("\n\n")

			// Show recent results
			start := 0
			if len(m.results) > 3 {
				start = len(m.results) - 3
			}

			for i := start; i < len(m.results); i++ {
				result := m.results[i]
				if result.Error != nil {
					content.WriteString(fmt.Sprintf("%s %s: %s\n",
						errorStyle.Render("✗"),
						result.Name,
						result.Error.Error()))
				} else {
					duration := result.Duration.Truncate(time.Millisecond)
					content.WriteString(fmt.Sprintf("%s %s [%s@%s] - %v\n",
						successStyle.Render("✓"),
						result.Name,
						result.Branch,
						result.Remote,
						duration))
				}
			}
		}
	}

	if !m.done {
		content.WriteString("\n\n")
		content.WriteString(infoStyle.Render("Press 'q' or Ctrl+C to quit"))
	}

	return content.String()
}

func (m *Model) renderSummary() string {
	var content strings.Builder

	if len(m.repos) == 0 {
		content.WriteString(titleStyle.Render("git-herd"))
		content.WriteString("\n\n")
		content.WriteString(infoStyle.Render(fmt.Sprintf("No Git repositories found in %s", m.rootPath)))
		return content.String()
	}

	// Header
	content.WriteString(titleStyle.Render("🎉 git-herd Results"))
	content.WriteString("\n\n")

	// Results
	successful := 0
	failed := 0

	for _, result := range m.results {
		if result.Error != nil {
			failed++
			if strings.Contains(result.Error.Error(), "skipped") {
				content.WriteString(fmt.Sprintf("%s %s (%s): %s\n",
					infoStyle.Render("⊝"),
					result.Name,
					result.Path,
					result.Error.Error()))
			} else {
				content.WriteString(fmt.Sprintf("%s %s (%s): %s\n",
					errorStyle.Render("✗"),
					result.Name,
					result.Path,
					result.Error.Error()))
			}
		} else {
			successful++
			status := "✓"
			if m.config.DryRun {
				status = "👁"
			}
			duration := result.Duration.Truncate(time.Millisecond)
			content.WriteString(fmt.Sprintf("%s %s (%s) [%s@%s] - %v\n",
				successStyle.Render(status),
				result.Name,
				result.Path,
				result.Branch,
				result.Remote,
				duration))
		}
	}

	// Count skipped separately from failed
	skipped := 0
	actualFailed := 0
	for _, result := range m.results {
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "skipped") {
				skipped++
			} else {
				actualFailed++
			}
		}
	}

	// Summary box
	summaryText := fmt.Sprintf("📊 Summary: %s successful, %s failed, %s skipped, %s total",
		successStyle.Render(fmt.Sprintf("%d", successful)),
		errorStyle.Render(fmt.Sprintf("%d", actualFailed)),
		infoStyle.Render(fmt.Sprintf("%d", skipped)),
		infoStyle.Render(fmt.Sprintf("%d", len(m.results))))

	content.WriteString("\n")
	content.WriteString(summaryStyle.Render(summaryText))

	// Save report if requested
	if m.config.SaveReport != "" {
		if err := saveReport(m.config, m.results, successful, actualFailed, skipped); err == nil {
			content.WriteString(fmt.Sprintf("\n📄 Detailed report saved to: %s", m.config.SaveReport))
		}
	}

	return content.String()
}
