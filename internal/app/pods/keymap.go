package pods

import (
	"kubeui/internal/pkg/kubeui"

	"github.com/charmbracelet/bubbles/key"
)

// appKeyMap defines the keys that are handled at the top level in the application.
// These keys will be checked before passing along a msg to underlying components.
type appKeyMap struct {
	kubeui.GlobalKeyMap
	exitView        key.Binding
	selectNamespace key.Binding
	refreshPodList  key.Binding
}

// newAppKeyMap defines the actual key bindings and creates an appKeyMap.
func newAppKeyMap() *appKeyMap {
	return &appKeyMap{
		GlobalKeyMap: kubeui.NewGlobalKeyMap(),
		exitView: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Exit current view"),
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
