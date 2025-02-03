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
		return ExitViewMsg{}
	}
}

// PushViewMsg is used to navigate to a new view.
type PushViewMsg struct {
	// Id of the view.
	// This is chosen by the application.
	Id string
	// Indicates whether view.Init should be called.
	// The results of this depends on the view, but most of the time this will result in a reload of data.
	// If going back to a previous view without any changes then you probably want to set this to false.
	Initialize bool
}

// PushView navigates to a new view.
// It is up the the application to map the id to an actual view.
func PushView(id string, initialize bool) tea.Cmd {
	return func() tea.Msg {
		return PushViewMsg{
			Id:         id,
			Initialize: initialize,
		}
	}
}

// PopViewMsg is used to navigate to the previous view.
type PopViewMsg struct {
	Initialize bool
}

// PopView navigates to the previous view.
// It is basically used to simluate popups, or used to display an error view where you can choose to quit or go back to the previous view.
func PopView(initialize bool) tea.Cmd {
	return func() tea.Msg {
		return PopViewMsg{
			Initialize: initialize,
		}
	}
}
