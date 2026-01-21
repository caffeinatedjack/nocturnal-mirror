package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Tabs represents tab navigation bar.
type Tabs struct {
	tabs    []string
	current Tab
	keys    KeyMap
}

// Styles for tabs.
var (
	tabsInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	tabsActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Underline(true)

	tabsContainerStyle = lipgloss.NewStyle().
				PaddingBottom(1)
)

// NewTabs creates a new tabs component.
func NewTabs(keys KeyMap) *Tabs {
	return &Tabs{
		tabs:    TabNames(),
		current: TabOverview,
		keys:    keys,
	}
}

// SetCurrent sets current tab.
func (t *Tabs) SetCurrent(tab Tab) {
	if tab >= TabOverview && tab <= TabStats {
		t.current = tab
	}
}

// Current returns current tab.
func (t *Tabs) Current() Tab {
	return t.current
}

// Next moves to next tab.
func (t *Tabs) Next() {
	if t.current < TabStats {
		t.current++
	} else {
		t.current = TabOverview
	}
}

// Prev moves to previous tab.
func (t *Tabs) Prev() {
	if t.current > TabOverview {
		t.current--
	} else {
		t.current = TabStats
	}
}

// View renders tabs.
func (t *Tabs) View() string {
	var tabs []string

	for i, tabName := range t.tabs {
		tab := Tab(i)
		if tab == t.current {
			tabs = append(tabs, tabsActiveStyle.Render(tabName))
		} else {
			tabs = append(tabs, tabsInactiveStyle.Render(tabName))
		}
	}

	return tabsContainerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, tabs...))
}
