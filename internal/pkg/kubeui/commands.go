package kubeui

import tea "github.com/charmbracelet/bubbletea"

// Exit exits the application.
func Exit() tea.Cmd {
	return tea.Quit
}

// Error is used to return errors.
func Error(err error) tea.Cmd {
	return func() tea.Msg { return err }
}

// ExitViewMsg is sent by a view when it wants to close.
type ExitViewMsg struct{}

// ExitViewMsg exits the current view.
func ExitView() tea.Cmd {
	return func() tea.Msg {
		return &ExitViewMsg{}
	}
}

// PushViewMsg is used to navigate to a new view.
type PushViewMsg struct {
	Id string
}

// PushView navigates to a new view.
// It is up the the application to map the id to an actual view.
func PushView(id string) tea.Cmd {
	return func() tea.Msg {
		return PushViewMsg{
			Id: id,
		}
	}
}
