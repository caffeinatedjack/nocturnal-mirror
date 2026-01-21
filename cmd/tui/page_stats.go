package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatsPage is the statistics page.
type StatsPage struct {
	width    int
	height   int
	content  string
	specPath string
}

// NewStatsPage creates a new stats page.
func NewStatsPage(specPath string) *StatsPage {
	return &StatsPage{
		content:  "Loading statistics...",
		specPath: specPath,
	}
}

// LoadData loads data for the stats page.
func (p *StatsPage) LoadData(specPath string) {
	p.specPath = specPath

	var lines []string

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	lines = append(lines, titleStyle.Render("📈 Project Statistics"))
	lines = append(lines, "")

	// Count proposals and their states
	proposalCount := 0
	proposalsPath := filepath.Join(specPath, "proposal")
	if entries, err := os.ReadDir(proposalsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				proposalCount++
			}
		}
	}

	// Count files in each directory
	ruleCount := 0
	rulesPath := filepath.Join(specPath, "rule")
	if files, err := os.ReadDir(rulesPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				ruleCount++
			}
		}
	}

	specCount := 0
	specsPath := filepath.Join(specPath, "section")
	if files, err := os.ReadDir(specsPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				specCount++
			}
		}
	}

	maintCount := 0
	maintPath := filepath.Join(specPath, "maintenance")
	if files, err := os.ReadDir(maintPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				maintCount++
			}
		}
	}

	docCount := 0
	docsPath := filepath.Join(specPath, "third")
	if files, err := os.ReadDir(docsPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				docCount++
			}
		}
	}

	archivedCount := 0
	archivePath := filepath.Join(specPath, "archive")
	if files, err := os.ReadDir(archivePath); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				archivedCount++
			}
		}
	}

	lines = append(lines, labelStyle.Render("Content Counts:"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s %s", labelStyle.Render("Proposals:"), valueStyle.Render(fmt.Sprintf("%d", proposalCount))))
	lines = append(lines, fmt.Sprintf("  %s %s", labelStyle.Render("Rules:"), valueStyle.Render(fmt.Sprintf("%d", ruleCount))))
	lines = append(lines, fmt.Sprintf("  %s %s", labelStyle.Render("Completed Specs:"), valueStyle.Render(fmt.Sprintf("%d", specCount))))
	lines = append(lines, fmt.Sprintf("  %s %s", labelStyle.Render("Maintenance Items:"), valueStyle.Render(fmt.Sprintf("%d", maintCount))))
	lines = append(lines, fmt.Sprintf("  %s %s", labelStyle.Render("Documentation:"), valueStyle.Render(fmt.Sprintf("%d", docCount))))
	lines = append(lines, fmt.Sprintf("  %s %s", labelStyle.Render("Archived:"), valueStyle.Render(fmt.Sprintf("%d", archivedCount))))
	lines = append(lines, "")

	total := proposalCount + ruleCount + specCount + maintCount + docCount
	lines = append(lines, fmt.Sprintf("%s %s", titleStyle.Render("Total Documents:"), valueStyle.Render(fmt.Sprintf("%d", total))))

	p.content = strings.Join(lines, "\n")
}

// SetSize sets the page size.
func (p *StatsPage) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles messages for the stats page.
func (p *StatsPage) Update(msg interface{}, model *Model) tea.Cmd {
	return nil
}

// View renders the stats page.
func (p *StatsPage) View() string {
	style := lipgloss.NewStyle().
		Padding(2, 4)

	return style.Width(p.width).Render(p.content)
}
