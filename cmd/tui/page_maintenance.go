package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MaintenancePage is the maintenance items page.
type MaintenancePage struct {
	width    int
	height   int
	detail   *Detail
	specPath string
	items    []ListItem
}

// NewMaintenancePage creates a new maintenance page.
func NewMaintenancePage(specPath string) *MaintenancePage {
	return &MaintenancePage{
		detail:   NewDetail(0),
		specPath: specPath,
		items:    []ListItem{},
	}
}

// LoadData loads data for the maintenance page.
func (p *MaintenancePage) LoadData(specPath string) {
	p.specPath = specPath
	p.items = []ListItem{}

	maintPath := filepath.Join(specPath, "maintenance")
	files, err := os.ReadDir(maintPath)
	if err != nil {
		if os.IsNotExist(err) {
			p.items = append(p.items, ListItem{
				ID:     "none",
				Title:  "No maintenance directory found",
				Status: "pending",
			})
		} else {
			p.items = append(p.items, ListItem{
				ID:     "error",
				Title:  fmt.Sprintf("Error reading maintenance: %v", err),
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
			filePath := filepath.Join(maintPath, file.Name())
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
				Status:   "pending",
			})
		}
	}

	if len(p.items) == 0 {
		p.items = append(p.items, ListItem{
			ID:     "none",
			Title:  "No maintenance items found",
			Status: "pending",
		})
	}

	p.detail.SetItems(p.items)
}

// SetSize sets the page size.
func (p *MaintenancePage) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.detail.SetHeight(height)
}

// Update handles messages for the maintenance page.
func (p *MaintenancePage) Update(msg interface{}, model *Model) tea.Cmd {
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
				// Load maintenance content
				maintPath := filepath.Join(p.specPath, "maintenance", item.ID+".md")
				if data, err := os.ReadFile(maintPath); err == nil {
					content := RenderMarkdown(string(data), p.width)
					p.detail.SetContent(content)
					p.detail.leftList.Select()
				}
			}
		case "e":
			// Open in external editor
			if item := p.detail.Selected(); item != nil && item.ID != "none" && item.ID != "error" {
				maintPath := filepath.Join(p.specPath, "maintenance", item.ID+".md")
				return OpenEditor(maintPath)
			}
		case "esc":
			// Deselect to go back to list navigation
			p.detail.leftList.ClearSelection()
		}
	}
	return nil
}

// View renders the maintenance page.
func (p *MaintenancePage) View() string {
	return p.detail.View(p.width)
}
