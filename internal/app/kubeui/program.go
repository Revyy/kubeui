package kubeui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func NewProgram(model tea.Model) *tea.Program {

	lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

	p := tea.NewProgram(model)

	return p
}

func StartProgram(p *tea.Program) {
	p.Start()
}
