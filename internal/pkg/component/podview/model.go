package podview

import (
	"kubeui/internal/pkg/component/columntable"
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
	keys         *KeyMap
	cursor       int
	sections     []string
	sectionViews map[string]podInfoFunc

	windowWidth int
	pod         v1.Pod
}

// Returns a list of keybindings to be used in help text.
func (pv Model) KeyList() []key.Binding {
	keyList := []key.Binding{
		pv.keys.Left,
		pv.keys.Right,
	}

	return keyList
}

type podInfoFunc func(pod v1.Pod, width int) string

// New creates a new Model.
func New(pod v1.Pod, windowWidth int) Model {

	return Model{
		keys:        newKeyMap(),
		windowWidth: windowWidth,
		pod:         pod,
		sections:    []string{"Status", "Annotations", "Labels"},
		sectionViews: map[string]podInfoFunc{
			"Status": func(pod v1.Pod, width int) string {
				columns, row := podStatusTable(pod)
				return lipgloss.NewStyle().Width(width).Render(columnTableData(columns, []*columntable.Row{row}))
			},
			"Annotations": func(pod v1.Pod, width int) string {
				columns, rows := stringMapTable("Key", "Value", pod.Annotations)
				return lipgloss.NewStyle().Width(width).Render(columnTableData(columns, rows))
			},
			"Labels": func(pod v1.Pod, width int) string {
				columns, rows := stringMapTable("Key", "Value", pod.Labels)
				return lipgloss.NewStyle().Width(width).Render(columnTableData(columns, rows))
			},
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

	tabSelect := lipgloss.NewStyle().Width(pv.windowWidth).Render(tabsBuilder.String())

	var mainBuilder strings.Builder

	mainBuilder.WriteString(tabSelect)
	mainBuilder.WriteString("\n\n")

	mainBuilder.WriteString(pv.sectionViews[pv.sections[pv.cursor]](pv.pod, pv.windowWidth))

	return mainBuilder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (n Model) Init() tea.Cmd {
	return nil
}
