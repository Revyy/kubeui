package podview

import (
	"encoding/json"
	"fmt"
	"kubeui/internal/pkg/jsoncolor"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/kubeui"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
	"github.com/muesli/reflow/wrap"
	"k8s.io/utils/integer"
)

// keyMap defines the key bindings for the PodView.
type keyMap struct {
	Left    key.Binding
	Right   key.Binding
	Refresh key.Binding
}

// newKeyMap creates a new KeyMap.
func newKeyMap() *keyMap {

	return &keyMap{
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "Move cursor left one position"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "Move cursor right one position"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "Refresh pod information"),
		),
	}
}

// Refresh signals that the data for the pod should be refreshed.
type Refresh struct {
	// Name of the pod.
	PodName string
}

// NewPod contains a new pod to replace the old one, most likely it is an updated version of the same pod.
type NewPod struct {
	// New pod
	Pod k8s.Pod
}

// Model defines a component that can view and query different parts of a kubernetes pod.
type Model struct {
	keys   *keyMap
	cursor int
	views  []string
	view   view

	annotationsViewPort viewport.Model
	labelsViewPort      viewport.Model
	eventsViewPort      viewport.Model
	logsViewPort        viewport.Model

	verticalMargin int
	windowWidth    int
	windowHeight   int
	pod            k8s.Pod
}

// Returns a list of keybindings to be used in help text.
func (pv Model) KeyList() []key.Binding {

	viewPortKeys := viewport.DefaultKeyMap()

	keyList := []key.Binding{
		pv.keys.Refresh,
		pv.keys.Left,
		pv.keys.Right,
		viewPortKeys.Up,
		viewPortKeys.Down,
		viewPortKeys.PageUp,
		viewPortKeys.PageDown,
		viewPortKeys.HalfPageUp,
		viewPortKeys.HalfPageDown,
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
	// EVENTS is used to display the latest events for the pod.
	EVENTS
	// LOGS is used to display the logs of the pod.
	LOGS
)

// stringToSelectedView maps a string to a view.
var stringToSelectedView = map[string]view{
	STATUS.String():      STATUS,
	ANNOTATIONS.String(): ANNOTATIONS,
	LABELS.String():      LABELS,
	EVENTS.String():      EVENTS,
	LOGS.String():        LOGS,
}

// String implements the stringer interface for view.
func (s view) String() string {
	switch s {
	case STATUS:
		return "STATUS"
	case ANNOTATIONS:
		return "ANNOTATIONS"
	case LABELS:
		return "LABELS"
	case EVENTS:
		return "EVENTS"
	case LOGS:
		return "LOGS"
	}
	return "UNKNOWN"
}

// New creates a new Model.
func New(pod k8s.Pod, verticalMargin, windowWidth, windowHeight int) Model {
	model := Model{
		keys:           newKeyMap(),
		windowWidth:    windowWidth,
		windowHeight:   windowHeight,
		verticalMargin: verticalMargin,
		pod:            pod,
		views:          []string{STATUS.String(), ANNOTATIONS.String(), LABELS.String(), EVENTS.String(), LOGS.String()},
	}

	model = model.updateViewportSizes()
	model = model.updateViewportContents()

	return model
}

func buildJSONLines(maxWidth int, jsonStr string) []string {

	formatter := jsoncolor.NewFormatter()

	return slices.Filter(slices.Map(strings.Split(jsonStr, "\n"), func(str string) string {
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(str), &obj)
		if err != nil {
			return ""
		}

		s, err := formatter.Marshal(obj)

		if err != nil {
			return ""
		}

		return wrap.String(string(s), maxWidth)
	}), func(s string) bool {
		return len(s) > 0
	})

}

// SetVerticalMargin sets a new vertical margin, this is used to calculate the height of the viewport.
// A parent component should call this if its content height prior to this components view changes.
func (pv Model) SetVerticalMargin(verticalMargin int) Model {
	pv.verticalMargin = verticalMargin
	pv = pv.updateViewportSizes()
	return pv
}

func (pv Model) updateViewportSizes() Model {
	pv.annotationsViewPort.Width = pv.windowWidth
	pv.annotationsViewPort.Height = pv.windowHeight - pv.calculateViewportOfset(ANNOTATIONS)

	pv.labelsViewPort.Width = pv.windowWidth
	pv.labelsViewPort.Height = pv.windowHeight - pv.calculateViewportOfset(LABELS)

	pv.eventsViewPort.Width = pv.windowWidth
	pv.eventsViewPort.Height = pv.windowHeight - pv.calculateViewportOfset(EVENTS)

	pv.logsViewPort.Width = pv.windowWidth
	pv.logsViewPort.Height = pv.windowHeight - pv.calculateViewportOfset(LOGS)

	return pv
}

func (pv Model) updateViewportContents() Model {
	pv.annotationsViewPort.SetContent(kubeui.RowsString(kubeui.StringMapTable(pv.windowWidth, "Key", "Value", pv.pod.Pod.Annotations)))
	pv.labelsViewPort.SetContent(kubeui.RowsString(kubeui.StringMapTable(pv.windowWidth, "Key", "Value", pv.pod.Pod.Labels)))
	pv.eventsViewPort.SetContent(kubeui.RowsString(kubeui.EventsTable(pv.windowWidth, pv.pod.Events)))

	pv.logsViewPort.SetContent(strings.Join(buildJSONLines(pv.windowWidth, pv.pod.Logs), "\n\n"))
	pv.logsViewPort.GotoBottom()

	return pv
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (pv Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// TODO: update viewports.
		pv.windowWidth = msg.Width
		pv.windowHeight = msg.Height
		pv = pv.updateViewportSizes()
		pv = pv.updateViewportContents()
		return pv, nil

	// Is it a key press?
	case tea.KeyMsg:
		switch {
		// The "left" key move the cursor left
		case key.Matches(msg, pv.keys.Left):
			if pv.cursor > 0 {
				pv.cursor--
			} else {
				pv.cursor = len(pv.views) - 1
			}
			pv.view = stringToSelectedView[pv.views[pv.cursor]]
			return pv, nil

		// The "right" key move the cursor right
		case key.Matches(msg, pv.keys.Right):
			if pv.cursor < len(pv.views)-1 {
				pv.cursor++
			} else {
				pv.cursor = 0
			}
			pv.view = stringToSelectedView[pv.views[pv.cursor]]
			return pv, nil

		case key.Matches(msg, pv.keys.Refresh):
			return pv, func() tea.Msg {
				return Refresh{
					PodName: pv.pod.Pod.GetName(),
				}
			}
		}
	case NewPod:
		// TODO: Set content of viewports again.
		pv.pod = msg.Pod
		pv = pv.updateViewportContents()
		return pv, nil
	}

	var cmd tea.Cmd

	switch pv.view {
	case ANNOTATIONS:
		pv.annotationsViewPort, cmd = pv.annotationsViewPort.Update(msg)
	case LABELS:
		pv.labelsViewPort, cmd = pv.labelsViewPort.Update(msg)
	case EVENTS:
		pv.eventsViewPort, cmd = pv.eventsViewPort.Update(msg)
	case LOGS:
		pv.logsViewPort, cmd = pv.logsViewPort.Update(msg)
	}

	return pv, cmd
}

// calculateViewportOfSet calculates the amount of space occupied by other components/views so that the viewPort can occupy the rest of the space.
func (pv Model) calculateViewportOfset(v view) int {
	return lipgloss.Height(pv.tableHeaderView(v)) + lipgloss.Height(pv.tabsView()) + lipgloss.Height(pv.footerView()) + pv.verticalMargin
}

// tabsView builds the tab select view.
func (pv Model) tabsView() string {
	return kubeui.TabsSelect(pv.cursor, pv.windowWidth, pv.views) + "\n\n"
}

// tableHeaderView creates the table header view.
// Producing table headers seperately from the rows allows us to let the content scroll past the headers without hiding them.
func (pv Model) tableHeaderView(v view) string {

	var columns []*kubeui.DataColumn
	switch v {
	case STATUS:
		columns, _ = kubeui.PodStatusTable(pv.pod.Pod)
	case ANNOTATIONS:
		columns, _ = kubeui.StringMapTable(pv.windowWidth, "Key", "Value", pv.pod.Pod.Annotations)
	case LABELS:
		columns, _ = kubeui.StringMapTable(pv.windowWidth, "Key", "Value", pv.pod.Pod.Labels)
	case EVENTS:
		columns, _ = kubeui.EventsTable(pv.eventsViewPort.Width, pv.pod.Events)
	case LOGS:
		return strings.Repeat("─", pv.logsViewPort.Width) + "\n"
	}

	line := strings.Repeat("─", pv.windowWidth)
	return lipgloss.NewStyle().Width(pv.windowWidth).Render(kubeui.ColumnsString(columns)) + "\n" + lipgloss.JoinHorizontal(lipgloss.Center, line) + "\n\n"
}

// footerView creates the footerView which contains information about how far the user has scrolled through the viewPort.
func (pv Model) footerView() string {

	var info string

	switch pv.view {
	case ANNOTATIONS:
		info = fmt.Sprintf("%3.f%%", pv.annotationsViewPort.ScrollPercent()*100)
	case LABELS:
		info = fmt.Sprintf("%3.f%%", pv.labelsViewPort.ScrollPercent()*100)
	case EVENTS:
		info = fmt.Sprintf("%3.f%%", pv.eventsViewPort.ScrollPercent()*100)
	case LOGS:
		info = fmt.Sprintf("%3.f%%", pv.logsViewPort.ScrollPercent()*100)
	}

	line := strings.Repeat("─", integer.IntMax(0, pv.windowWidth-lipgloss.Width(info)))
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (pv Model) View() string {

	var builder strings.Builder
	builder.WriteString(pv.tabsView())
	builder.WriteString(pv.tableHeaderView(pv.view))

	switch pv.view {
	case STATUS:
		columns, row := kubeui.PodStatusTable(pv.pod.Pod)
		builder.WriteString(kubeui.RowsString(columns, []*kubeui.DataRow{row}))
	case ANNOTATIONS:
		builder.WriteString(pv.annotationsViewPort.View())
	case LABELS:
		builder.WriteString(pv.labelsViewPort.View())
	case EVENTS:
		builder.WriteString(pv.eventsViewPort.View())
	case LOGS:
		builder.WriteString(pv.logsViewPort.View())
	}

	// STATUS does not have a footer.
	if pv.view != STATUS {
		builder.WriteString(pv.footerView())
	}

	return builder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (pv Model) Init() tea.Cmd {
	return nil
}
