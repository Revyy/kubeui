package pods

import "github.com/charmbracelet/bubbles/key"

// appKeyMap defines the keys that are handled at the top level in the application.
// These keys will be checked before passing along a msg to underlying components.
type appKeyMap struct {
	quit            key.Binding
	exitView        key.Binding
	help            key.Binding
	selectNamespace key.Binding
	refreshPodList  key.Binding
}

// newAppKeyMap defines the actual key bindings and creates an appKeyMap.
func newAppKeyMap() *appKeyMap {
	return &appKeyMap{
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c,ctrl+q", "Quit"),
		),
		exitView: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Exit current view"),
		),
		help: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "Toggle help"),
		),
		selectNamespace: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "Select namespace"),
		),
		refreshPodList: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "Refresh pod list"),
		),
	}
}
