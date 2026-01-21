package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// List represents a generic list component with selection.
type ListItem struct {
	ID       string
	Title    string
	Subtitle string
	Status   string // "active", "completed", "due", "pending", etc.
}

// List component.
type List struct {
	items    []ListItem
	cursor   int
	selected int
	viewport *viewport.Model
	height   int
	styles   ListStyles
}

// ListStyles holds styling for the list.
type ListStyles struct {
	Cursor         lipgloss.Style
	Selected       lipgloss.Style
	Item           lipgloss.Style
	StatusActive   lipgloss.Style
	StatusComplete lipgloss.Style
	StatusDue      lipgloss.Style
	StatusPending  lipgloss.Style
	Dim            lipgloss.Style
}

// DefaultListStyles returns default list styles.
func DefaultListStyles() ListStyles {
	return ListStyles{
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),
		StatusActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
		StatusComplete: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
		StatusDue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")),
		StatusPending: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}

// NewList creates a new list component.
func NewList(height int) *List {
	vp := viewport.New(0, height)
	vp.GotoTop()

	return &List{
		items:    make([]ListItem, 0),
		cursor:   0,
		selected: -1,
		viewport: &vp,
		height:   height,
		styles:   DefaultListStyles(),
	}
}

// SetItems sets the list items.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.cursor = 0
	l.selected = -1
	l.viewport.GotoTop()
}

// SetStyles sets custom styles.
func (l *List) SetStyles(styles ListStyles) {
	l.styles = styles
}

// Cursor returns the current cursor position.
func (l *List) Cursor() int {
	return l.cursor
}

// Selected returns the selected item.
func (l *List) Selected() *ListItem {
	if l.selected >= 0 && l.selected < len(l.items) {
		return &l.items[l.selected]
	}
	if l.cursor >= 0 && l.cursor < len(l.items) {
		return &l.items[l.cursor]
	}
	return nil
}

// MoveUp moves the cursor up.
func (l *List) MoveUp() {
	if len(l.items) == 0 {
		return
	}
	if l.cursor > 0 {
		l.cursor--
		l.viewport.LineUp(1)
	}
}

// MoveDown moves the cursor down.
func (l *List) MoveDown() {
	if len(l.items) == 0 {
		return
	}
	if l.cursor < len(l.items)-1 {
		l.cursor++
		l.viewport.LineDown(1)
	}
}

// Select selects the current item.
func (l *List) Select() {
	l.selected = l.cursor
}

// ClearSelection clears the selection.
func (l *List) ClearSelection() {
	l.selected = -1
}

// IsSelected checks if an item is selected.
func (l *List) IsSelected() bool {
	return l.selected >= 0
}

// View renders the list.
func (l *List) View() string {
	if len(l.items) == 0 {
		return l.styles.Dim.Render("No items")
	}

	var lines []string
	for i, item := range l.items {
		line := l.renderItem(i, item)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	l.viewport.SetContent(content)
	return l.viewport.View()
}

// renderItem renders a single list item.
func (l *List) renderItem(index int, item ListItem) string {
	var prefix string
	var style lipgloss.Style

	if index == l.selected {
		prefix = "● "
		style = l.styles.Selected
	} else if index == l.cursor {
		prefix = "→ "
		style = l.styles.Item
	} else {
		prefix = "  "
		style = l.styles.Item
	}

	// Add status indicator
	status := ""
	switch item.Status {
	case "active":
		status = l.styles.StatusActive.Render("[active]")
	case "completed":
		status = l.styles.StatusComplete.Render("[✓]")
	case "due":
		status = l.styles.StatusDue.Render("[!]")
	case "pending":
		status = l.styles.StatusPending.Render("[ ]")
	}

	if status != "" {
		status = " " + status
	}

	line := prefix + style.Render(item.Title) + status

	if item.Subtitle != "" {
		line += "\n  " + l.styles.Dim.Render(item.Subtitle)
	}

	return line
}

// SetHeight sets the list height.
func (l *List) SetHeight(height int) {
	l.height = height
	l.viewport.Height = height
}

// SyncViewport syncs the viewport with the current cursor position.
func (l *List) SyncViewport() {
	if l.cursor < 0 || l.cursor >= len(l.items) {
		return
	}

	// Calculate visible range
	visibleHeight := l.viewport.Height
	topLine := l.viewport.YOffset
	bottomLine := topLine + visibleHeight

	if l.cursor < topLine {
		l.viewport.GotoTop()
		l.viewport.LineDown(l.cursor)
	} else if l.cursor >= bottomLine {
		l.viewport.GotoBottom()
		l.viewport.LineUp(len(l.items) - l.cursor - 1)
	}
}
