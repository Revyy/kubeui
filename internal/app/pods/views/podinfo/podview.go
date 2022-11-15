package podinfo

import (
	"encoding/json"
	"fmt"
	"kubeui/internal/pkg/jsoncolor"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/k8scommand"
	"kubeui/internal/pkg/kubeui"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
	"github.com/muesli/reflow/wrap"
	"k8s.io/utils/integer"
)

// keyMap defines the keys that are handled by this view.
type keyMap struct {
	kubeui.GlobalKeyMap
	Left         key.Binding
	Right        key.Binding
	NumberChoice key.Binding
}

// newKeyMap defines the actual key bindings and creates a keyMap.
func newKeyMap() *keyMap {
	return &keyMap{
		GlobalKeyMap: kubeui.NewGlobalKeyMap(),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "Move cursor left one position"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "Move cursor right one position"),
		),
		NumberChoice: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
			key.WithHelp("1,2,3,4,5,6,7,8,9", "Select container"),
		),
	}
}

func (v View) fullHelp() [][]key.Binding {

	bindings := [][]key.Binding{
		{v.keys.Help, v.keys.Quit, v.keys.Refresh, v.keys.ExitView},
	}

	viewPortKeys := viewport.DefaultKeyMap()

	bindings = append(bindings, []key.Binding{
		v.keys.Refresh,
		v.keys.Left,
		v.keys.Right,
		v.keys.NumberChoice,
		viewPortKeys.Up,
		viewPortKeys.Down,
		viewPortKeys.PageUp,
		viewPortKeys.PageDown,
		viewPortKeys.HalfPageUp,
		viewPortKeys.HalfPageDown,
	})

	return bindings
}

// New creates a new View.
func New(c kubeui.Context) View {

	return View{
		keys: newKeyMap(),
		tabs: []string{STATUS.String(), ANNOTATIONS.String(), LABELS.String(), EVENTS.String(), LOGS.String()},
	}
}

// View displays pod information.
type View struct {
	keys *keyMap

	// Cursor keeps track of which tab is active.
	cursor int
	tab    tab
	tabs   []string

	// List of container names.
	selectedContainer string
	containerNames    []string

	// Viewports for scrolling content
	annotationsViewPort viewport.Model
	labelsViewPort      viewport.Model
	eventsViewPort      viewport.Model
	logsViewPort        viewport.Model

	// Show full help view or not.
	showFullHelp bool

	pod *k8s.Pod
}

// tab defines the different tabs of the component.
type tab uint16

const (
	// STATUS is used to display status information about the pod.
	STATUS tab = iota
	// ANNOTATIONS is used to display the annotations set for the pod.
	ANNOTATIONS
	// LABELS is used to display the labels set for the pod.
	LABELS
	// EVENTS is used to display the latest events for the pod.
	EVENTS
	// LOGS is used to display the logs of the pod.
	LOGS
)

// stringToSelectedTab maps a string to a tab.
var stringToSelectedTab = map[string]tab{
	STATUS.String():      STATUS,
	ANNOTATIONS.String(): ANNOTATIONS,
	LABELS.String():      LABELS,
	EVENTS.String():      EVENTS,
	LOGS.String():        LOGS,
}

// String implements the stringer interface for tab.
func (t tab) String() string {
	switch t {
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

// Update handles new messages from the runtime.
func (v View) Update(c kubeui.Context, msg kubeui.Msg) (kubeui.Context, kubeui.View, tea.Cmd) {

	if msg.IsKeyMsg() && v.showFullHelp {
		v.showFullHelp = false
		return c, v, nil
	}

	// Keys
	switch {
	case msg.MatchesKeyBindings(v.keys.Quit):
		return c, v, kubeui.Exit()

	case msg.MatchesKeyBindings(v.keys.ExitView):
		return c, v, kubeui.PushView("pod_selection", false)

	case msg.MatchesKeyBindings(v.keys.Help) && !v.showFullHelp:
		v.showFullHelp = true
		return c, v, nil

	case msg.MatchesKeyBindings(v.keys.Left):
		v = v.moveTabLeft()
		return c, v, nil
	case msg.MatchesKeyBindings(v.keys.Right):
		v = v.moveTabRight()
		return c, v, nil
	case msg.MatchesKeyBindings(v.keys.Refresh):
		return c, v, k8scommand.GetPod(c.Kubectl, c.Namespace, c.SelectedPod)
	case msg.MatchesKeyBindings(v.keys.NumberChoice) && v.tab == LOGS:

		v, err := v.selectContainer(msg)

		if err != nil {
			return c, v, kubeui.Error(err)
		}

		// If the selected container has logs then update the logview.
		if _, ok := v.pod.Logs[v.selectedContainer]; ok {
			v.logsViewPort.SetContent(strings.Join(buildJSONLines(c.WindowWidth, v.pod.Logs[v.selectedContainer]), "\n\n") + "\n\n")
			v.logsViewPort.GotoBottom()
		}

		return c, v, nil
	}

	// Results
	switch t := msg.TeaMsg.(type) {
	case k8scommand.GetPodMsg:
		v.pod = t.Pod
		v.containerNames = v.pod.ContainerNames()

		// If we don't have a selected container
		if v.selectedContainer == "" && len(v.containerNames) > 0 {
			v.selectedContainer = v.containerNames[0]
		}

		// If the selected container has logs then update the logview.
		if _, ok := t.Pod.Logs[v.selectedContainer]; ok {
			v.logsViewPort.SetContent(strings.Join(buildJSONLines(c.WindowWidth, t.Pod.Logs[v.selectedContainer]), "\n\n") + "\n\n")
		}

		v.annotationsViewPort.SetContent(kubeui.RowsString(kubeui.StringMapTable(c.WindowWidth, "Key", "Value", v.pod.Pod.Annotations)))
		v.labelsViewPort.SetContent(kubeui.RowsString(kubeui.StringMapTable(c.WindowWidth, "Key", "Value", v.pod.Pod.Labels)))
		v.eventsViewPort.SetContent(kubeui.RowsString(kubeui.EventsTable(c.WindowWidth, v.pod.Events)))

		for _, viewPort := range []*viewport.Model{&v.annotationsViewPort, &v.labelsViewPort, &v.eventsViewPort, &v.logsViewPort} {
			viewPort.Height = c.WindowHeight - lipgloss.Height(v.headerView(c.WindowWidth)) + lipgloss.Height(footerView(c.WindowWidth, *viewPort))
		}

		v.logsViewPort.GotoBottom()

		return c, v, nil
	}

	// Update viewports.
	var cmd tea.Cmd
	v, cmd = v.updateViewports(msg.TeaMsg)

	return c, v, cmd
}

// buildJSONLines builds colored json log lines.
func buildJSONLines(maxWidth int, jsonStr string) []string {

	formatter := jsoncolor.NewFormatter()
	formatter.StringMaxLength = maxWidth * 10

	return slices.Filter(slices.Map(strings.Split(jsonStr, "\n"), func(str string) string {
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(str), &obj)
		if err != nil {
			return wrap.String(str, maxWidth)
		}

		s, err := formatter.Marshal(obj)

		if err != nil {
			return wrap.String(str, maxWidth)
		}

		return wrap.String(string(s), maxWidth)
	}), func(s string) bool {
		return len(s) > 0
	})

}

// updateViewports updates the currently active viewport.
func (v View) updateViewports(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch v.tab {
	case ANNOTATIONS:
		v.annotationsViewPort, cmd = v.annotationsViewPort.Update(msg)
	case LABELS:
		v.labelsViewPort, cmd = v.labelsViewPort.Update(msg)
	case EVENTS:
		v.eventsViewPort, cmd = v.eventsViewPort.Update(msg)
	case LOGS:
		v.logsViewPort, cmd = v.logsViewPort.Update(msg)
	}

	return v, cmd
}

func (v View) moveTabLeft() View {
	if v.cursor > 0 {
		v.cursor--
	} else {
		v.cursor = len(v.tabs) - 1
	}
	v.tab = stringToSelectedTab[v.tabs[v.cursor]]

	return v
}

func (v View) moveTabRight() View {
	if v.cursor < len(v.tabs)-1 {
		v.cursor++
	} else {
		v.cursor = 0
	}
	v.tab = stringToSelectedTab[v.tabs[v.cursor]]

	return v
}

func (v View) selectContainer(msg kubeui.Msg) (View, error) {
	key := msg.TeaMsg.(tea.KeyMsg)
	intKey, err := strconv.Atoi(key.String())

	if err != nil {
		return v, err
	}
	// Subtract one to make it match the index for container names.
	intKey--

	if intKey >= 0 && intKey <= len(v.containerNames)-1 {
		v.selectedContainer = v.containerNames[intKey]
	}

	return v, nil
}

// View renders the ui of the view.
func (v View) View(c kubeui.Context) string {

	if v.showFullHelp {
		return kubeui.FullHelp(c.WindowWidth, v.fullHelp())
	}

	builder := strings.Builder{}
	header := v.headerView(c.WindowWidth)
	builder.WriteString(header)

	if v.pod == nil {
		return builder.String()
	}

	switch v.tab {
	case STATUS:
		columns, row := kubeui.PodStatusTable(v.pod.Pod)
		builder.WriteString(kubeui.RowsString(columns, []*kubeui.DataRow{row}))
		return builder.String()

	case ANNOTATIONS:
		footer := footerView(c.WindowWidth, v.annotationsViewPort)
		offset := lipgloss.Height(header) + lipgloss.Height(footer)
		builder.WriteString(renderViewport(c.WindowWidth, c.WindowHeight, offset, v.annotationsViewPort))
		builder.WriteString(footer)

	case LABELS:
		footer := footerView(c.WindowWidth, v.labelsViewPort)
		offset := lipgloss.Height(header) + lipgloss.Height(footer)
		builder.WriteString(renderViewport(c.WindowWidth, c.WindowHeight, offset, v.labelsViewPort))
		builder.WriteString(footer)

	case EVENTS:
		footer := footerView(c.WindowWidth, v.labelsViewPort)
		offset := lipgloss.Height(header) + lipgloss.Height(footer)
		builder.WriteString(renderViewport(c.WindowWidth, c.WindowHeight, offset, v.eventsViewPort))
		builder.WriteString(footer)

	case LOGS:
		footer := footerView(c.WindowWidth, v.logsViewPort)

		offset := lipgloss.Height(header) + lipgloss.Height(footer) //+ lipgloss.Height(containers)
		builder.WriteString(renderViewport(c.WindowWidth, c.WindowHeight, offset, v.logsViewPort))
		builder.WriteString(footer)
	}

	return builder.String()
}

func (v View) headerView(width int) string {
	if v.pod == nil {
		return "Loading..."
	}

	builder := strings.Builder{}

	builder.WriteString(kubeui.ShortHelp(width, []key.Binding{
		v.keys.Help,
		v.keys.Quit,
		v.keys.Refresh,
		v.keys.Left,
		v.keys.Right,
	}))

	builder.WriteString("\n\n")

	builder.WriteString(kubeui.TabsSelect(v.cursor, width, v.tabs) + "\n\n")

	if v.tab == LOGS {
		builder.WriteString(kubeui.HorizontalSelectList(v.containerNames, v.selectedContainer, width))
		builder.WriteString("\n")
	}

	builder.WriteString(tableHeaderView(width, v.tab, *v.pod))

	return builder.String()
}

func renderViewport(windowWidth, windowHeight int, offset int, viewPort viewport.Model) string {
	viewPort.Width = windowWidth
	viewPort.Height = windowHeight - offset
	return viewPort.View()
}

// tableHeaderView creates the table header view.
// Producing table headers seperately from the rows allows us to let the content scroll past the headers without hiding them.
func tableHeaderView(width int, t tab, pod k8s.Pod) string {

	var columns []*kubeui.DataColumn
	switch t {
	case STATUS:
		columns, _ = kubeui.PodStatusTable(pod.Pod)
	case ANNOTATIONS:
		columns, _ = kubeui.StringMapTable(width, "Key", "Value", pod.Pod.Annotations)
	case LABELS:
		columns, _ = kubeui.StringMapTable(width, "Key", "Value", pod.Pod.Labels)
	case EVENTS:
		columns, _ = kubeui.EventsTable(width, pod.Events)
	case LOGS:
		return strings.Repeat("─", width) + "\n"
	}

	line := strings.Repeat("─", width)
	return lipgloss.NewStyle().Width(width).Render(kubeui.ColumnsString(columns)) + "\n" + lipgloss.JoinHorizontal(lipgloss.Center, line) + "\n\n"
}

// footerView creates the footerView which contains information about how far the user has scrolled through the viewPort.
func footerView(width int, viewPort viewport.Model) string {

	info := fmt.Sprintf("%3.f%%", viewPort.ScrollPercent()*100)
	line := strings.Repeat("─", integer.IntMax(0, width-lipgloss.Width(info)))
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// Init initializes the view.
func (v View) Init(c kubeui.Context) tea.Cmd {
	return k8scommand.GetPod(c.Kubectl, c.Namespace, c.SelectedPod)
}

// Destroy is called before a view is removed as the active view in the application.
func (v View) Destroy(c kubeui.Context) tea.Cmd {
	return nil
}
