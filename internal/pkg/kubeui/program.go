package kubeui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// NewProgram creates a new bubbletea program given a bubbletea model.
func NewProgram(model tea.Model, useAltScreen bool) *tea.Program {

	// Needed as lipgloss uses stdout/stdin to communicate with the terminal to check if it has a dark or light background
	// Once the bubbletea program starts it takes control of stdout and stdin.
	lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

	options := []tea.ProgramOption{}

	if useAltScreen {
		options = append(options, tea.WithAltScreen())
	}

	p := tea.NewProgram(model, options...)

	return p
}

// StartProgram starts a given bubbletea program.
func StartProgram(p *tea.Program) {
	p.Start()
}
