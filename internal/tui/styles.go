package tui

import "github.com/charmbracelet/lipgloss"

var (
	// One Dark Pro Flat Theme Colors
	// Source: https://github.com/Binaryify/OneDark-Pro
	oneDarkBlue  = lipgloss.Color("#4aa5f0")
	oneDarkGreen = lipgloss.Color("#8cc265")
	oneDarkRed   = lipgloss.Color("#e05561")
	oneDarkGrey  = lipgloss.Color("#5c6370")
	// oneDarkYellow = lipgloss.Color("#d18f52")
	// oneDarkFg     = lipgloss.Color("#abb2bf")

	// App Styles
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(oneDarkBlue)
	categoryStyle = lipgloss.NewStyle().Bold(true).Foreground(oneDarkBlue)
	cursorStyle   = lipgloss.NewStyle().Foreground(oneDarkBlue)

	dimStyle = lipgloss.NewStyle().Foreground(oneDarkGrey)
	okStyle  = lipgloss.NewStyle().Foreground(oneDarkGreen)
	badStyle = lipgloss.NewStyle().Foreground(oneDarkRed)
)
