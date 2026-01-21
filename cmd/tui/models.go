package tui

// Tab represents a TUI page tab.
type Tab int

const (
	TabOverview Tab = iota
	TabProposals
	TabRules
	TabMaintenance
	TabDocs
	TabConfig
	TabStats
)

func (t Tab) String() string {
	switch t {
	case TabOverview:
		return "Overview"
	case TabProposals:
		return "Proposals"
	case TabRules:
		return "Rules"
	case TabMaintenance:
		return "Maintenance"
	case TabDocs:
		return "Docs"
	case TabConfig:
		return "Config"
	case TabStats:
		return "Stats"
	default:
		return "Unknown"
	}
}

// TabNames returns all tab names.
func TabNames() []string {
	return []string{
		TabOverview.String(),
		TabProposals.String(),
		TabRules.String(),
		TabMaintenance.String(),
		TabDocs.String(),
		TabConfig.String(),
		TabStats.String(),
	}
}

// TabSelectMsg is sent when a tab is selected.
type TabSelectMsg struct {
	Tab
}

// ShowHelpMsg is sent to toggle help display.
type ShowHelpMsg struct {
	Show bool
}

// RefreshMsg is sent to refresh data.
type RefreshMsg struct{}

// EditorDoneMsg is sent after external editor completes.
type EditorDoneMsg struct {
	Path string
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Err error
}

// SuccessMsg is sent for successful operations.
type SuccessMsg struct {
	Message string
}
