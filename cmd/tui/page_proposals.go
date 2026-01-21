package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// ProposalsPage is the proposals management page.
type ProposalsPage struct {
	width    int
	height   int
	detail   *Detail
	specPath string
	items    []ListItem
}

// NewProposalsPage creates a new proposals page.
func NewProposalsPage(specPath string) *ProposalsPage {
	return &ProposalsPage{
		detail:   NewDetail(0),
		specPath: specPath,
		items:    []ListItem{},
	}
}

// LoadData loads data for the proposals page.
func (p *ProposalsPage) LoadData(specPath string) {
	p.specPath = specPath
	p.items = []ListItem{}

	proposalsPath := filepath.Join(specPath, "proposal")
	entries, err := os.ReadDir(proposalsPath)
	if err != nil {
		if os.IsNotExist(err) {
			p.items = append(p.items, ListItem{
				ID:     "none",
				Title:  "No proposals directory found",
				Status: "pending",
			})
		} else {
			p.items = append(p.items, ListItem{
				ID:     "error",
				Title:  fmt.Sprintf("Error reading proposals: %v", err),
				Status: "pending",
			})
		}
		p.detail.SetItems(p.items)
		return
	}

	// Get active proposal
	activeSlug := getActiveProposal(specPath)

	for _, entry := range entries {
		if entry.IsDir() {
			slug := entry.Name()
			status := "pending"
			if slug == activeSlug {
				status = "active"
			}

			// Get proposal path
			proposalPath := filepath.Join(proposalsPath, slug)

			// Check if implementation.md exists
			implPath := filepath.Join(proposalPath, "implementation.md")
			subtitle := ""
			if _, err := os.Stat(implPath); err == nil {
				subtitle = "Has implementation.md"
			}

			p.items = append(p.items, ListItem{
				ID:       slug,
				Title:    slug,
				Subtitle: subtitle,
				Status:   status,
			})
		}
	}

	if len(p.items) == 0 {
		p.items = append(p.items, ListItem{
			ID:     "none",
			Title:  "No proposals found",
			Status: "pending",
		})
	}

	p.detail.SetItems(p.items)
}

// SetSize sets the page size.
func (p *ProposalsPage) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.detail.SetHeight(height)
}

// Update handles messages for the proposals page.
func (p *ProposalsPage) Update(msg interface{}, model *Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.detail.leftList.IsSelected() {
				// If content is showing, scroll up
				p.detail.ScrollUp()
			} else {
				// Otherwise navigate list
				p.detail.MoveUp()
			}
		case "down", "j":
			if p.detail.leftList.IsSelected() {
				// If content is showing, scroll down
				p.detail.ScrollDown()
			} else {
				// Otherwise navigate list
				p.detail.MoveDown()
			}
		case "enter":
			// Select and show content
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				// Load proposal content
				proposalPath := filepath.Join(p.specPath, "proposal", item.ID)
				implPath := filepath.Join(proposalPath, "implementation.md")
				if data, err := os.ReadFile(implPath); err == nil {
					content := RenderMarkdown(string(data), p.width)
					p.detail.SetContent(content)
					p.detail.leftList.Select()
				}
			}
		case "e":
			// Open in external editor
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				proposalPath := filepath.Join(p.specPath, "proposal", item.ID)
				implPath := filepath.Join(proposalPath, "implementation.md")
				return OpenEditor(implPath)
			}
		case "a":
			// Activate proposal
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				return ActivateProposal(p.specPath, item.ID)
			}
		case "c":
			// Complete proposal
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				return CompleteProposal(p.specPath, item.ID)
			}
		case "v":
			// Validate proposal
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				return ValidateProposal(p.specPath, item.ID)
			}
		case "d":
			// Delete proposal (no force for safety)
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				return DeleteProposal(p.specPath, item.ID, false)
			}
		case "x":
			// Deactivate proposal
			return DeactivateProposal(p.specPath)
		case "esc":
			// Deselect to go back to list navigation
			p.detail.leftList.ClearSelection()
		}
	}
	return nil
}

// View renders the proposals page.
func (p *ProposalsPage) View() string {
	return p.detail.View(p.width)
}
