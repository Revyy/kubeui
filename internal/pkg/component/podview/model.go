package podview

import (
	"kubeui/internal/pkg/component/columntable"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v1 "k8s.io/api/core/v1"
)

// keyMap defines the key bindings for the PodView.
type keyMap struct {
	Left  key.Binding
	Right key.Binding
}

// newKeyMap creates a new KeyMap.
func newKeyMap() *keyMap {

	return &keyMap{
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
	keys   *keyMap
	cursor int
	views  []string
	view   view

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

// view defines the different views of the component.
type view uint16

const (
	// STATUS is used to display status information about the pod.
	STATUS view = iota
	// ANNOTATIONS is used to display the annotations set for the pod.
	ANNOTATIONS
	// LABELS is used to display the labels set for the pod.
	LABELS
	// LOGS is used to display the logs of the pod.
	LOGS
)

var stringToSelectedView = map[string]view{
	STATUS.String():      STATUS,
	ANNOTATIONS.String(): ANNOTATIONS,
	LABELS.String():      LABELS,
	LOGS.String():        LOGS,
}

func (s view) String() string {
	switch s {
	case STATUS:
		return "STATUS"
	case ANNOTATIONS:
		return "ANNOTATIONS"
	case LABELS:
		return "LABELS"
	case LOGS:
		return "LOGS"
	}
	return "UNKNOWN"
}

// New creates a new Model.
func New(pod v1.Pod, windowWidth int) Model {

	return Model{
		keys:        newKeyMap(),
		windowWidth: windowWidth,
		pod:         pod,
		views:       []string{STATUS.String(), ANNOTATIONS.String(), LABELS.String()},
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
				pv.cursor = len(pv.views) - 1
			}
			pv.view = stringToSelectedView[pv.views[pv.cursor]]

			return pv, nil

		// The "down" key move the cursor down
		case key.Matches(msg, pv.keys.Right):
			if pv.cursor < len(pv.views)-1 {
				pv.cursor++
			} else {
				pv.cursor = 0
			}
			pv.view = stringToSelectedView[pv.views[pv.cursor]]

			return pv, nil
		}
	}
	return pv, nil
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (pv Model) View() string {

	var builder strings.Builder

	builder.WriteString(tabsBuilder(pv.cursor, pv.windowWidth, pv.views))
	builder.WriteString("\n\n")

	windowWithStyle := lipgloss.NewStyle().Width(pv.windowWidth)

	switch pv.view {
	case STATUS:
		columns, row := podStatusTable(pv.pod)
		builder.WriteString(windowWithStyle.Render(columnTableData(columns, []*columntable.Row{row})))
	case ANNOTATIONS:
		columns, rows := stringMapTable("Key", "Value", pv.pod.Annotations)
		builder.WriteString(windowWithStyle.Render(columnTableData(columns, rows)))
	case LABELS:
		columns, rows := stringMapTable("Key", "Value", pv.pod.Labels)
		builder.WriteString(windowWithStyle.Render(columnTableData(columns, rows)))
	}

	return builder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (n Model) Init() tea.Cmd {
	return nil
}
