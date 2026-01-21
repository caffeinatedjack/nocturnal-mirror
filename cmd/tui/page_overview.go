package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OverviewPage is the overview dashboard page.
type OverviewPage struct {
	width    int
	height   int
	content  string
	specPath string
}

// NewOverviewPage creates a new overview page.
func NewOverviewPage(specPath string) *OverviewPage {
	return &OverviewPage{
		content:  "Loading overview...",
		specPath: specPath,
	}
}

// LoadData loads data for the overview page.
func (p *OverviewPage) LoadData(specPath string) {
	p.specPath = specPath

	var lines []string

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	lines = append(lines, titleStyle.Render("📊 Project Overview"))
	lines = append(lines, "")

	// Count proposals
	proposalCount := 0
	proposalsPath := filepath.Join(specPath, "proposal")
	if entries, err := os.ReadDir(proposalsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				proposalCount++
			}
		}
	}

	// Count rules
	ruleCount := 0
	rulesPath := filepath.Join(specPath, "rule")
	if files, err := os.ReadDir(rulesPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				ruleCount++
			}
		}
	}

	// Count completed specs
	specCount := 0
	specsPath := filepath.Join(specPath, "section")
	if files, err := os.ReadDir(specsPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				specCount++
			}
		}
	}

	// Count maintenance items
	maintCount := 0
	maintPath := filepath.Join(specPath, "maintenance")
	if files, err := os.ReadDir(maintPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				maintCount++
			}
		}
	}

	// Count docs
	docCount := 0
	docsPath := filepath.Join(specPath, "third")
	if files, err := os.ReadDir(docsPath); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				docCount++
			}
		}
	}

	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Proposals:"), valueStyle.Render(fmt.Sprintf("%d", proposalCount))))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Rules:"), valueStyle.Render(fmt.Sprintf("%d", ruleCount))))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Completed Specs:"), valueStyle.Render(fmt.Sprintf("%d", specCount))))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Maintenance Items:"), valueStyle.Render(fmt.Sprintf("%d", maintCount))))
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("Documentation:"), valueStyle.Render(fmt.Sprintf("%d", docCount))))
	lines = append(lines, "")

	// Check for active proposal
	activeSlug := getActiveProposal(specPath)
	if activeSlug != "" {
		lines = append(lines, titleStyle.Render("📋 Active Proposal"))
		lines = append(lines, "")
		lines = append(lines, valueStyle.Render(activeSlug))

		// Try to read proposal description
		proposalPath := filepath.Join(specPath, "proposal", activeSlug, "implementation.md")
		if data, err := os.ReadFile(proposalPath); err == nil {
			// Get first line or heading
			content := string(data)
			firstLines := strings.Split(content, "\n")
			if len(firstLines) > 0 {
				desc := strings.TrimPrefix(firstLines[0], "# ")
				if len(desc) > 0 && len(desc) < 100 {
					lines = append(lines, "")
					lines = append(lines, labelStyle.Render(desc))
				}
			}
		}
	} else {
		lines = append(lines, titleStyle.Render("📋 No Active Proposal"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Use 'nocturnal spec proposal activate <name>' to activate a proposal"))
	}

	p.content = strings.Join(lines, "\n")
}

// SetSize sets the page size.
func (p *OverviewPage) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles messages for the overview page.
func (p *OverviewPage) Update(msg interface{}, model *Model) tea.Cmd {
	return nil
}

// View renders the overview page.
func (p *OverviewPage) View() string {
	style := lipgloss.NewStyle().
		Padding(2, 4)

	return style.Width(p.width).Render(p.content)
}
