package podview

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// KeyMap defines the key bindings for the PodView.
type KeyMap struct {
	Exit key.Binding
}

// newKeyMap creates a new KeyMap.
func newKeyMap() *KeyMap {

	return &KeyMap{
		Exit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Exit podview"),
		),
	}
}

// Model defines a component that can view and query different parts of a kubernetes pod.
type Model struct {
	keys     *KeyMap
	cursor   int
	sections []string

	// ClientSet used to issue commands to kubernetes.
	kubectl *kubernetes.Clientset
	pod     *v1.Pod
}

// Returns a list of keybindings to be used in help text.
func (pv Model) KeyList() []key.Binding {
	keyList := []key.Binding{
		pv.keys.Exit,
	}

	return keyList
}

// New creates a new Model.
func New() Model {
	return Model{
		keys: newKeyMap(),
	}
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (pv Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return pv, nil
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (n Model) View() string {
	return ""
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (n Model) Init() tea.Cmd {
	return nil
}
