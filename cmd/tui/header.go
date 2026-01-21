package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Header represents always-visible header bar.
type Header struct {
	version        string
	specPath       string
	activeProposal string
}

// NewHeader creates a new header.
func NewHeader(version, specPath, activeProposal string) *Header {
	return &Header{
		version:        version,
		specPath:       specPath,
		activeProposal: activeProposal,
	}
}

// Styles for header.
var (
	headerStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15"))

	versionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	activeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	headerDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

// View renders header.
func (h *Header) View(width int) string {
	activePart := ""
	if h.activeProposal != "" {
		activePart = headerDimStyle.Render(" | ") + activeStyle.Render("[active: "+h.activeProposal+"]")
	}

	left := versionStyle.Render("nocturnal v" + h.version)
	middle := pathStyle.Render(h.specPath)
	right := activePart

	// Calculate widths
	leftWidth := lipgloss.Width(left)
	middleWidth := lipgloss.Width(middle)
	rightWidth := lipgloss.Width(right)
	separatorWidth := lipgloss.Width(headerDimStyle.Render(" | "))

	totalWidth := leftWidth + separatorWidth + middleWidth
	if rightWidth > 0 {
		totalWidth += separatorWidth + rightWidth
	}

	// Build header
	if totalWidth <= width {
		// Everything fits
		return headerStyle.Width(width).Render(
			left +
				headerDimStyle.Render(" | ") +
				middle +
				right,
		)
	}

	// Need to truncate middle
	availableWidth := width - leftWidth - (separatorWidth * 2)
	if rightWidth > 0 {
		availableWidth -= separatorWidth + rightWidth
	}

	if availableWidth < 10 {
		// Middle too small, just show left and active
		return headerStyle.Width(width).Render(
			left +
				headerDimStyle.Render(" | ") +
				right,
		)
	}

	middle = pathStyle.MaxWidth(availableWidth).Render(h.specPath)
	return headerStyle.Width(width).Render(
		left +
			headerDimStyle.Render(" | ") +
			middle +
			headerDimStyle.Render(" | ") +
			right,
	)
}

// UpdateActiveProposal updates active proposal display.
func (h *Header) UpdateActiveProposal(proposal string) {
	h.activeProposal = proposal
}
