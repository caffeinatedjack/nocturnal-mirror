package tui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the main TUI model.
type Model struct {
	keys       KeyMap
	currentTab Tab
	tabs       *Tabs
	header     *Header
	status     *Status
	viewport   *viewport.Model

	// Page models
	overviewPage    *OverviewPage
	proposalsPage   *ProposalsPage
	rulesPage       *RulesPage
	maintenancePage *MaintenancePage
	docsPage        *DocsPage
	configPage      *ConfigPage
	statsPage       *StatsPage

	// Other
	specPath  string
	quitting  bool
	lastError string
	watcher   *Watcher
}

// Init initializes the TUI model.
func (m Model) Init() bubbletea.Cmd {
	// Load initial data
	m.overviewPage.LoadData(m.specPath)
	m.proposalsPage.LoadData(m.specPath)
	m.rulesPage.LoadData(m.specPath)
	m.maintenancePage.LoadData(m.specPath)
	m.docsPage.LoadData(m.specPath)
	m.configPage.LoadData(m.specPath)
	m.statsPage.LoadData(m.specPath)

	// Update header with active proposal
	activeSlug := getActiveProposal(m.specPath)
	if activeSlug != "" {
		m.header.UpdateActiveProposal(activeSlug)
	}

	return nil
}

// Update handles TUI messages.
func (m Model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var cmds []bubbletea.Cmd

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		// Check for quit
		if m.keys.IsQuitKey(msg) {
			if m.watcher != nil {
				_ = m.watcher.Close()
			}
			m.quitting = true
			return m, bubbletea.Quit
		}

		// Check for help
		if m.keys.IsHelpKey(msg) {
			m.status.ToggleHelp()
			return m, nil
		}

		// Check for refresh
		if m.keys.IsRefreshKey(msg) {
			m.status.SetInfo("Refreshing...")
			return m, bubbletea.Tick(time.Second*1, func(t time.Time) bubbletea.Msg {
				return RefreshMsg{}
			})
		}

		// Handle tab navigation
		if m.keys.IsLeftKey(msg) {
			m.tabs.Prev()
			m.currentTab = m.tabs.Current()
			return m, nil
		}
		if m.keys.IsRightKey(msg) {
			m.tabs.Next()
			m.currentTab = m.tabs.Current()
			return m, nil
		}

		// Delegate to current page
		var cmd bubbletea.Cmd
		switch m.currentTab {
		case TabOverview:
			cmd = m.overviewPage.Update(msg, &m)
		case TabProposals:
			cmd = m.proposalsPage.Update(msg, &m)
		case TabRules:
			cmd = m.rulesPage.Update(msg, &m)
		case TabMaintenance:
			cmd = m.maintenancePage.Update(msg, &m)
		case TabDocs:
			cmd = m.docsPage.Update(msg, &m)
		case TabConfig:
			cmd = m.configPage.Update(msg, &m)
		case TabStats:
			cmd = m.statsPage.Update(msg, &m)
		}

		if cmd != nil {
			return m, cmd
		}

	case bubbletea.WindowSizeMsg:
		// Update viewport size
		m.viewport.Width = msg.Width
		viewportHeight := msg.Height - 4 // header, tabs, status
		m.viewport.Height = viewportHeight

		// Update page sizes
		if m.overviewPage != nil {
			m.overviewPage.SetSize(msg.Width, viewportHeight)
		}
		if m.proposalsPage != nil {
			m.proposalsPage.SetSize(msg.Width, viewportHeight)
		}
		if m.rulesPage != nil {
			m.rulesPage.SetSize(msg.Width, viewportHeight)
		}
		if m.maintenancePage != nil {
			m.maintenancePage.SetSize(msg.Width, viewportHeight)
		}
		if m.docsPage != nil {
			m.docsPage.SetSize(msg.Width, viewportHeight)
		}
		if m.configPage != nil {
			m.configPage.SetSize(msg.Width, viewportHeight)
		}
		if m.statsPage != nil {
			m.statsPage.SetSize(msg.Width, viewportHeight)
		}

	case RefreshMsg:
		m.refreshData()
		return m, nil

	case EditorDoneMsg:
		m.status.SetSuccess("File saved")
		m.refreshData()
		return m, nil

	case ErrorMsg:
		m.status.SetError(msg.Err.Error())
		return m, nil

	case SuccessMsg:
		m.status.SetSuccess(msg.Message)
		m.refreshData()
		return m, nil

	case ShowHelpMsg:
		m.status.ToggleHelp()
		return m, nil
	}

	// Update status
	if cmd := m.status.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, bubbletea.Batch(cmds...)
}

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Render current page
	var pageView string
	switch m.currentTab {
	case TabOverview:
		pageView = m.overviewPage.View()
	case TabProposals:
		pageView = m.proposalsPage.View()
	case TabRules:
		pageView = m.rulesPage.View()
	case TabMaintenance:
		pageView = m.maintenancePage.View()
	case TabDocs:
		pageView = m.docsPage.View()
	case TabConfig:
		pageView = m.configPage.View()
	case TabStats:
		pageView = m.statsPage.View()
	}

	// Build full view
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(m.viewport.Width),
		m.tabs.View(),
		pageView,
		m.status.View(m.viewport.Width),
	)
}

// refreshData refreshes data for all pages.
func (m *Model) refreshData() {
	m.overviewPage.LoadData(m.specPath)
	m.proposalsPage.LoadData(m.specPath)
	m.rulesPage.LoadData(m.specPath)
	m.maintenancePage.LoadData(m.specPath)
	m.docsPage.LoadData(m.specPath)
	m.configPage.LoadData(m.specPath)
	m.statsPage.LoadData(m.specPath)
}

// NewModel creates a new TUI model.
func NewModel(specPath, version string) Model {
	keys := DefaultKeyMap()

	// Create components
	header := NewHeader(version, specPath, "")
	tabs := NewTabs(keys)
	status := NewStatus(keys)

	// Create pages
	overviewPage := NewOverviewPage(specPath)
	proposalsPage := NewProposalsPage(specPath)
	rulesPage := NewRulesPage(specPath)
	maintenancePage := NewMaintenancePage(specPath)
	docsPage := NewDocsPage(specPath)
	configPage := NewConfigPage(specPath)
	statsPage := NewStatsPage(specPath)

	// Create watcher
	watcher, _ := NewWatcher(specPath)

	// Create viewport
	vp := viewport.New(80, 24)

	return Model{
		keys:            keys,
		currentTab:      TabOverview,
		tabs:            tabs,
		header:          header,
		status:          status,
		viewport:        &vp,
		specPath:        specPath,
		overviewPage:    overviewPage,
		proposalsPage:   proposalsPage,
		rulesPage:       rulesPage,
		maintenancePage: maintenancePage,
		docsPage:        docsPage,
		configPage:      configPage,
		statsPage:       statsPage,
		watcher:         watcher,
	}
}

// Run starts the TUI.
func Run(specPath, version string) error {
	// Check if workspace exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("specification workspace not initialized. Run 'nocturnal spec init' first")
	}

	model := NewModel(specPath, version)
	p := bubbletea.NewProgram(
		model,
		bubbletea.WithAltScreen(),
		bubbletea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}

// Editor runs external editor (simplified version without tea dependency).
func EditorRun(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors
		editors := []string{"vim", "nvim", "vi", "nano", "code --wait"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
