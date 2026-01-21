package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// RulesPage is the rules management page.
type RulesPage struct {
	width    int
	height   int
	detail   *Detail
	specPath string
	items    []ListItem
}

// NewRulesPage creates a new rules page.
func NewRulesPage(specPath string) *RulesPage {
	return &RulesPage{
		detail:   NewDetail(0),
		specPath: specPath,
		items:    []ListItem{},
	}
}

// LoadData loads data for the rules page.
func (p *RulesPage) LoadData(specPath string) {
	p.specPath = specPath
	p.items = []ListItem{}

	rulesPath := filepath.Join(specPath, "rule")
	files, err := os.ReadDir(rulesPath)
	if err != nil {
		if os.IsNotExist(err) {
			p.items = append(p.items, ListItem{
				ID:     "none",
				Title:  "No rules directory found",
				Status: "pending",
			})
		} else {
			p.items = append(p.items, ListItem{
				ID:     "error",
				Title:  fmt.Sprintf("Error reading rules: %v", err),
				Status: "pending",
			})
		}
		p.detail.SetItems(p.items)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			name := strings.TrimSuffix(file.Name(), ".md")

			// Read first line for subtitle
			filePath := filepath.Join(rulesPath, file.Name())
			subtitle := ""
			if data, err := os.ReadFile(filePath); err == nil {
				lines := strings.Split(string(data), "\n")
				if len(lines) > 0 {
					firstLine := strings.TrimPrefix(lines[0], "# ")
					if len(firstLine) < 80 {
						subtitle = firstLine
					}
				}
			}

			p.items = append(p.items, ListItem{
				ID:       name,
				Title:    name,
				Subtitle: subtitle,
				Status:   "completed",
			})
		}
	}

	if len(p.items) == 0 {
		p.items = append(p.items, ListItem{
			ID:     "none",
			Title:  "No rules found",
			Status: "pending",
		})
	}

	p.detail.SetItems(p.items)
}

// SetSize sets the page size.
func (p *RulesPage) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.detail.SetHeight(height)
}

// Update handles messages for the rules page.
func (p *RulesPage) Update(msg interface{}, model *Model) tea.Cmd {
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
				// Load rule content
				rulePath := filepath.Join(p.specPath, "rule", item.ID+".md")
				if data, err := os.ReadFile(rulePath); err == nil {
					content := RenderMarkdown(string(data), p.width)
					p.detail.SetContent(content)
					p.detail.leftList.Select()
				}
			}
		case "e":
			// Open in external editor
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				rulePath := filepath.Join(p.specPath, "rule", item.ID+".md")
				return OpenEditor(rulePath)
			}
		case "esc":
			// Deselect to go back to list navigation
			p.detail.leftList.ClearSelection()
		}
	}
	return nil
}

// View renders the rules page.
func (p *RulesPage) View() string {
	return p.detail.View(p.width)
}
