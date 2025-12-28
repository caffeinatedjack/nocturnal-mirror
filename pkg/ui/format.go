// Package ui provides terminal UI formatting helpers.
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Red      = lipgloss.Color("9")
	Yellow   = lipgloss.Color("11")
	Green    = lipgloss.Color("10")
	Blue     = lipgloss.Color("12")
	DimColor = lipgloss.Color("8")

	// Styles
	ErrorStyle   = lipgloss.NewStyle().Foreground(Red)
	WarningStyle = lipgloss.NewStyle().Foreground(Yellow)
	SuccessStyle = lipgloss.NewStyle().Foreground(Green)
	InfoStyle    = lipgloss.NewStyle().Foreground(Blue)
	DimStyle     = lipgloss.NewStyle().Foreground(DimColor)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	TopicStyle   = lipgloss.NewStyle().Foreground(Blue).Bold(true)
)

// Success prints a success message.
func Success(msg string) {
	fmt.Println(SuccessStyle.Render(msg))
}

// Error prints an error message.
func Error(msg string) {
	fmt.Println(ErrorStyle.Render(msg))
}

// Warning prints a warning message.
func Warning(msg string) {
	fmt.Println(WarningStyle.Render(msg))
}

// Info prints an info message.
func Info(msg string) {
	fmt.Println(InfoStyle.Render(msg))
}

// PrintDim prints a dimmed message.
func PrintDim(msg string) {
	fmt.Println(DimStyle.Render(msg))
}
