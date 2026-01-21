package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Detail represents a two-pane detail view (list + content).
type Detail struct {
	leftList *List
	viewport *viewport.Model
	content  string
	height   int
	showLeft bool
	percent  int // percentage of width for left panel
}

// Styles for detail view.
var (
	detailBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8"))

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12"))

	detailDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

// NewDetail creates a new detail view.
func NewDetail(height int) *Detail {
	vp := viewport.New(0, height)
	vp.GotoTop()

	return &Detail{
		leftList: NewList(height),
		viewport: &vp,
		content:  "",
		height:   height,
		showLeft: true,
		percent:  30,
	}
}

// SetItems sets the left list items.
func (d *Detail) SetItems(items []ListItem) {
	d.leftList.SetItems(items)
}

// SetContent sets the right content.
func (d *Detail) SetContent(content string) {
	d.content = content
	d.viewport.SetContent(content)
	d.viewport.GotoTop()
}

// SetHeight sets the detail view height.
func (d *Detail) SetHeight(height int) {
	d.height = height
	d.leftList.SetHeight(height)
	d.viewport.Height = height
}

// SetSplit sets the split percentage for the left panel.
func (d *Detail) SetSplit(percent int) {
	d.percent = percent
	if d.percent < 10 {
		d.percent = 10
	}
	if d.percent > 50 {
		d.percent = 50
	}
}

// SetShowLeft toggles the left panel visibility.
func (d *Detail) SetShowLeft(show bool) {
	d.showLeft = show
}

// Selected returns the selected list item.
func (d *Detail) Selected() *ListItem {
	return d.leftList.Selected()
}

// MoveUp moves the selection up.
func (d *Detail) MoveUp() {
	if d.showLeft {
		d.leftList.MoveUp()
	}
}

// MoveDown moves the selection down.
func (d *Detail) MoveDown() {
	if d.showLeft {
		d.leftList.MoveDown()
	}
}

// View renders the detail view.
func (d *Detail) View(width int) string {
	if !d.showLeft {
		// Only show content
		return d.renderContent(width)
	}

	// Calculate widths
	leftWidth := (width * d.percent) / 100
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := width - leftWidth - 2 // 2 for border

	// Render left and right panels
	leftPanel := d.renderLeft(leftWidth)
	rightPanel := d.renderContent(rightWidth)

	return lipgloss.JoinHorizontal(lipgloss.Left, leftPanel, rightPanel)
}

// renderLeft renders the left list panel.
func (d *Detail) renderLeft(width int) string {
	leftContent := d.leftList.View()
	return detailBorderStyle.Width(width).Height(d.height).Render(leftContent)
}

// renderContent renders the right content panel.
func (d *Detail) renderContent(width int) string {
	d.viewport.Width = width
	return d.viewport.View()
}

// ScrollUp scrolls the content up.
func (d *Detail) ScrollUp() {
	d.viewport.LineUp(1)
}

// ScrollDown scrolls the content down.
func (d *Detail) ScrollDown() {
	d.viewport.LineDown(1)
}

// PageUp scrolls the content up a page.
func (d *Detail) PageUp() {
	halfPage := d.viewport.Height / 2
	for i := 0; i < halfPage; i++ {
		d.viewport.LineUp(1)
	}
}

// PageDown scrolls the content down a page.
func (d *Detail) PageDown() {
	halfPage := d.viewport.Height / 2
	for i := 0; i < halfPage; i++ {
		d.viewport.LineDown(1)
	}
}

// GotoTop scrolls to top.
func (d *Detail) GotoTop() {
	d.viewport.GotoTop()
}

// GotoBottom scrolls to bottom.
func (d *Detail) GotoBottom() {
	d.viewport.GotoBottom()
}

// RenderMarkdown renders markdown content with basic styling.
func RenderMarkdown(content string, width int) string {
	lines := strings.Split(content, "\n")
	var rendered []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			// Heading 1
			rendered = append(rendered, detailTitleStyle.Render(line))
		} else if strings.HasPrefix(line, "## ") {
			// Heading 2
			rendered = append(rendered, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render(line))
		} else if strings.HasPrefix(line, "### ") {
			// Heading 3
			rendered = append(rendered, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Render(line))
		} else if strings.HasPrefix(line, "- ") {
			// List item
			rendered = append(rendered, "  "+line)
		} else if strings.TrimSpace(line) == "---" {
			// Separator
			rendered = append(rendered, detailDimStyle.Render(strings.Repeat("─", width)))
		} else if strings.HasPrefix(line, "> ") {
			// Quote
			rendered = append(rendered, detailDimStyle.Render("  "+line))
		} else if strings.HasPrefix(line, "```") {
			// Code block - simple handling
			rendered = append(rendered, detailDimStyle.Render("  Code block"))
		} else {
			// Regular text
			rendered = append(rendered, line)
		}
	}

	return strings.Join(rendered, "\n")
}

// GetContentPreview returns a preview of content (first few lines).
func GetContentPreview(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	return strings.Join(lines[:maxLines], "\n") + "\n" + detailDimStyle.Render("...")
}

// FormatStatus formats status text with styling.
func FormatStatus(status string, styles ListStyles) string {
	switch status {
	case "active":
		return styles.StatusActive.Render("[active]")
	case "completed":
		return styles.StatusComplete.Render("[✓]")
	case "due":
		return styles.StatusDue.Render("[!]")
	case "pending":
		return styles.StatusPending.Render("[ ]")
	default:
		return ""
	}
}
