package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigPage is the configuration page.
type ConfigPage struct {
	width    int
	height   int
	content  string
	specPath string
}

// NewConfigPage creates a new config page.
func NewConfigPage(specPath string) *ConfigPage {
	return &ConfigPage{
		content:  "Loading configuration...",
		specPath: specPath,
	}
}

// LoadData loads data for the config page.
func (p *ConfigPage) LoadData(specPath string) {
	p.specPath = specPath

	configPath := filepath.Join(specPath, "nocturnal.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			p.content = "No configuration file found\n\nConfiguration file should be at:\n" + configPath
		} else {
			p.content = fmt.Sprintf("Error reading configuration:\n%v", err)
		}
		return
	}

	p.content = string(data)
}

// SetSize sets the page size.
func (p *ConfigPage) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles messages for the config page.
func (p *ConfigPage) Update(msg interface{}, model *Model) tea.Cmd {
	return nil
}

// View renders the config page.
func (p *ConfigPage) View() string {
	style := lipgloss.NewStyle().
		Padding(2, 4)

	return style.Width(p.width).Render(p.content)
}
