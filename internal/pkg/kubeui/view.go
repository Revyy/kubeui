package kubeui

import tea "github.com/charmbracelet/bubbletea"

// View is intended to be a stateful component that completely
// takes over the ui and handles most inputs except for some global keypresses and system messages.
type View interface {
	tea.Model
}

// ExitViewMsg is sent by a view when it wants to close.
type ExitViewMsg struct{}
