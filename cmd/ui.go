package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorRed    = lipgloss.Color("9")
	colorYellow = lipgloss.Color("11")
	colorGreen  = lipgloss.Color("10")
	colorBlue   = lipgloss.Color("12")
	colorDim    = lipgloss.Color("8")

	errorStyle   = lipgloss.NewStyle().Foreground(colorRed)
	warningStyle = lipgloss.NewStyle().Foreground(colorYellow)
	successStyle = lipgloss.NewStyle().Foreground(colorGreen)
	infoStyle    = lipgloss.NewStyle().Foreground(colorBlue)
	dimStyle     = lipgloss.NewStyle().Foreground(colorDim)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	topicStyle   = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
)

func printSuccess(msg string) {
	fmt.Println(successStyle.Render(msg))
}

func printError(msg string) {
	fmt.Println(errorStyle.Render(msg))
}

func printWarning(msg string) {
	fmt.Println(warningStyle.Render(msg))
}

func printInfo(msg string) {
	fmt.Println(infoStyle.Render(msg))
}

func printDim(msg string) {
	fmt.Println(dimStyle.Render(msg))
}
