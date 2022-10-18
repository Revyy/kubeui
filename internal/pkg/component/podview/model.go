package podview

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v1 "k8s.io/api/core/v1"
)

// KeyMap defines the key bindings for the PodView.
type KeyMap struct {
	Left  key.Binding
	Right key.Binding
}

// newKeyMap creates a new KeyMap.
func newKeyMap() *KeyMap {

	return &KeyMap{
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("up", "Move cursor left one position"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("down", "Move cursor right one position"),
		),
	}
}

// Model defines a component that can view and query different parts of a kubernetes pod.
type Model struct {
	keys     *KeyMap
	cursor   int
	sections []string

	windowWidth int

	pod v1.Pod
}

// Returns a list of keybindings to be used in help text.
func (pv Model) KeyList() []key.Binding {
	keyList := []key.Binding{
		pv.keys.Left,
		pv.keys.Right,
	}

	return keyList
}

// New creates a new Model.
func New(pod v1.Pod, windowWidth int) Model {
	return Model{
		keys:        newKeyMap(),
		windowWidth: windowWidth,
		sections: []string{
			"Status",
			"Annotations",
			"Labels",
		},
	}
}

// SetWindowWidth sets a new window width value for the podview.
func (pv Model) SetWindowWidth(width int) Model {
	pv.windowWidth = width
	return pv
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (pv Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch {
		// The "up" key move the cursor up
		case key.Matches(msg, pv.keys.Left):
			if pv.cursor > 0 {
				pv.cursor--
			} else {
				pv.cursor = len(pv.sections) - 1
			}
			return pv, nil

		// The "down" key move the cursor down
		case key.Matches(msg, pv.keys.Right):
			if pv.cursor < len(pv.sections)-1 {
				pv.cursor++
			} else {
				pv.cursor = 0
			}

		}
	}
	return pv, nil
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (pv Model) View() string {

	var tabsBuilder strings.Builder

	// Iterate over the items in the current page and print them out.
	for i, section := range pv.sections {

		// Is the cursor pointing at this choice?
		if pv.cursor == i {
			tabsBuilder.WriteString(lipgloss.NewStyle().Underline(true).Render(section) + " ")
			continue
		}

		tabsBuilder.WriteString(section + " ")
	}

	return lipgloss.NewStyle().Width(pv.windowWidth).Render(tabsBuilder.String())
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (n Model) Init() tea.Cmd {
	return nil
}
