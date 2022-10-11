package kubeui

import "github.com/charmbracelet/lipgloss"

// ErrorMessageStyle is used to style error messages.
var ErrorMessageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9"))

// StatusBarStyle is used to create a status bar displaying key information about the running app.
var StatusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("201")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	BorderBottom(true)
