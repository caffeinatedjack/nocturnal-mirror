package tui

import (
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Status represents status bar at the bottom.
type Status struct {
	message     string
	messageType string // "info", "error", "success"
	showHelp    bool
	autoDismiss bool
}

// Styles for status.
var (
	statusContainerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("8")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	statusInfoStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("12")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	statusErrorStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("9")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	statusSuccessStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("10")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))
)

// NewStatus creates a new status bar.
func NewStatus(keys KeyMap) *Status {
	return &Status{
		autoDismiss: true,
	}
}

// SetMessage sets a status message.
func (s *Status) SetMessage(msg string, msgType string) {
	s.message = msg
	s.messageType = msgType
}

// SetError sets an error message.
func (s *Status) SetError(msg string) {
	s.SetMessage(msg, "error")
}

// SetSuccess sets a success message.
func (s *Status) SetSuccess(msg string) {
	s.SetMessage(msg, "success")
}

// SetInfo sets an info message.
func (s *Status) SetInfo(msg string) {
	s.SetMessage(msg, "info")
}

// Clear clears the current message.
func (s *Status) Clear() {
	s.message = ""
	s.messageType = ""
}

// ToggleHelp toggles help display.
func (s *Status) ToggleHelp() {
	s.showHelp = !s.showHelp
}

// View renders the status bar.
func (s *Status) View(width int) string {
	if s.showHelp {
		helpText := helpStyle.Render(
			"Navigation: ↑↓/jk | Tabs: ←→/hl | Enter:view | e:edit | Esc:back | ?:help | r:refresh | q:quit",
		)
		return statusContainerStyle.Width(width).Render(helpText)
	}

	// Show message if present
	if s.message != "" {
		var style lipgloss.Style
		switch s.messageType {
		case "error":
			style = statusErrorStyle
		case "success":
			style = statusSuccessStyle
		case "info":
			style = statusInfoStyle
		default:
			style = statusContainerStyle
		}
		return statusContainerStyle.Width(width).Render(style.Render(s.message))
	}

	// Default status
	defaultText := helpStyle.Render("Press ? for help")
	return statusContainerStyle.Width(width).Render(defaultText)
}

// Update handles status messages.
func (s *Status) Update(msg bubbletea.Msg) bubbletea.Cmd {
	switch msg := msg.(type) {
	case ShowHelpMsg:
		s.showHelp = msg.Show
	case ErrorMsg:
		s.SetError(msg.Err.Error())
		if s.autoDismiss {
			return waitForDismiss(5)
		}
	case SuccessMsg:
		s.SetSuccess(msg.Message)
		if s.autoDismiss {
			return waitForDismiss(3)
		}
	case clearMsg:
		s.Clear()
	}
	return nil
}

// waitForDismiss creates a command to clear message after duration.
func waitForDismiss(seconds int) bubbletea.Cmd {
	return bubbletea.Tick(0, func(t time.Time) bubbletea.Msg {
		return clearMsg{}
	})
}

type clearMsg struct{}
