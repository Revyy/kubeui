package errorview

import tea "github.com/charmbracelet/bubbletea"

// New creates a new ErrorView.
func New(message string) ErrorView {
	return ErrorView{
		keys:    newKeyMap(),
		message: message,
	}
}

// ErrorView displays an error and allows the user to quit the app.
type ErrorView struct {
	keys    *keyMap
	message string
}

func (e ErrorView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return e, nil
}

func (e ErrorView) View() string {
	return ""
}

func (e ErrorView) Init() tea.Cmd {
	return nil
}
