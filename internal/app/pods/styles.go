package pods

import "github.com/charmbracelet/lipgloss"

var statusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("201")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	BorderBottom(true)

var errorMessageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9"))
