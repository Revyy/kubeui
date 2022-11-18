package kubeui

import (
	"github.com/charmbracelet/lipgloss"
)

// ErrorMessageStyle is used to style error messages.
var ErrorMessageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9"))

// StatusBarStyle is used to create a status bar displaying key information about the running app.
var StatusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("201")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	BorderBottom(true)

// SelectedStyle is used for selected items.
var SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})

// UnselectedStyle is used for unselected items.
var UnselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})
