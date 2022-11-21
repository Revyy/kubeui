package kubeui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Msg wraps the bubbletea message in a way that allows us to simplify some things.
type Msg struct {
	TeaMsg tea.Msg
}

// IsError tries to extract an error.
func (m Msg) IsError() (error, bool) {
	if e, ok := m.TeaMsg.(error); ok {
		return e, ok
	}
	return nil, false
}

// IsWindowResize checks if the msg contains a tea.WindowResizeMsg.
func (m Msg) IsWindowResize() bool {
	_, ok := m.TeaMsg.(tea.WindowSizeMsg)
	return ok
}

// GetWindowResizeMsg tries to extract a tea.WindowSizeMsg from the msg.
func (m Msg) GetWindowResizeMsg() (tea.WindowSizeMsg, bool) {
	w, ok := m.TeaMsg.(tea.WindowSizeMsg)
	return w, ok
}

// IsKeyMsg checks if the message contains a key click.
func (m Msg) IsKeyMsg() bool {

	_, ok := m.TeaMsg.(tea.KeyMsg)

	return ok
}

// MatchesKeyBindings checks if the message matches a specific KeyBinding.
func (m Msg) MatchesKeyBindings(bindings ...key.Binding) bool {

	keyMsg, ok := m.TeaMsg.(tea.KeyMsg)

	if !ok {
		return false
	}

	return key.Matches(keyMsg, bindings...)
}
