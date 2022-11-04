package kubeui

import (
	"github.com/charmbracelet/bubbles/key"
)

// GlobalKeyMap defines the keys that should be processed no matter which view is active in an application.
type GlobalKeyMap struct {
	Quit key.Binding
	Help key.Binding
}

// NewGlobalKeyMap defines the actual key bindings and creates a GlobalKeyMap.
func NewGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c,ctrl+q", "Quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "Toggle help"),
		),
	}
}
