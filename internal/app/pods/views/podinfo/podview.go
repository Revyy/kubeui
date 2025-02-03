package podinfo

import (
	"fmt"
	"strconv"
	"strings"

	"kubeui/internal/pkg/jsoncolor"
	"kubeui/internal/pkg/k8s/pods"
	"kubeui/internal/pkg/k8smsg"
	"kubeui/internal/pkg/kubeui"
	"kubeui/internal/pkg/ui/help"
	"kubeui/internal/pkg/ui/selection"
	"kubeui/internal/pkg/ui/table"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// K8sService represents the interface towards kubernetes needed by this view.
type K8sService interface {
	GetPod(namespace, id string) (*pods.Pod, error)
}

// View displays pod information.
type View struct {
	keys *keyMap

	tab  tab
	tabs []string

	// Indicates whether the pod has been loaded or not.
	initialized bool

	// List of container names.
	selectedContainer string
	containerNames    []string

	// Viewports for scrolling content
	annotationsViewPort viewport.Model
	labelsViewPort      viewport.Model
	eventsViewPort      viewport.Model
	logsViewPort        viewport.Model

	windowWidth  int
	windowHeight int

	// Show full help view or not.
	showFullHelp bool

	pod *pods.Pod

	// Kubernetes client.
	k8sClient K8sService
}

// New creates a new View.
func New(k8sClient K8sService, windowWidth, windowHeight int) View {
	return View{
		k8sClient:    k8sClient,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
		keys:         newKeyMap(),
		tabs:         []string{STATUS.String(), ANNOTATIONS.String(), LABELS.String(), EVENTS.String(), LOGS.String()},
	}
}

// tab defines the different tabs of the component.
type tab int

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

	case msg.IsWindowResize():
		windowResizeMsg, ok := msg.GetWindowResizeMsg()

		if !ok {
			return c, v, nil
		}

		v.windowHeight = windowResizeMsg.Height
		v.windowWidth = windowResizeMsg.Width
		v = v.updateViewportsAfterResize()
		return c, v, nil

	case msg.MatchesKeyBindings(v.keys.Quit):
		return c, v, kubeui.Exit()

	case msg.MatchesKeyBindings(v.keys.ExitView):
		return c, v, kubeui.PopView(false)

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

		pod, err := v.k8sClient.GetPod(c.Namespace, c.SelectedPod)
		if err != nil {
			return c, v, kubeui.Error(err)
		}

		return c, v, kubeui.GenericCmd(k8smsg.NewGetPodMsg(pod))
	case msg.MatchesKeyBindings(v.keys.NumberChoice) && v.tab == LOGS:

		v, err := v.selectContainer(msg)
		if err != nil {
			return c, v, kubeui.Error(err)
		}

		// If the selected container has logs then update the logview.
		if _, ok := v.pod.Logs[v.selectedContainer]; ok {
			v.logsViewPort.SetContent(strings.Join(jsoncolor.JSONLines(v.windowWidth, v.pod.Logs[v.selectedContainer]), "\n\n"))
			v.logsViewPort.GotoBottom()
		}

		return c, v, nil
	}

	// Results
	switch t := msg.TeaMsg.(type) {
	case k8smsg.GetPodMsg:

		if !v.initialized {
			v.initialized = true
		}

		v.pod = t.Pod
		v.containerNames = v.pod.ContainerNames()

		// If we don't have a selected container
		if v.selectedContainer == "" && len(v.containerNames) > 0 {
			v.selectedContainer = v.containerNames[0]
		}

		v = v.updateViewportsAfterResize()

		return c, v, nil
	}

	// Update viewports.
	var cmd tea.Cmd
	if v.initialized {
		v, cmd = v.updateViewports(msg.TeaMsg)
	}

	return c, v, cmd
}

func (v View) updateViewportsAfterResize() View {
	v.annotationsViewPort.Height = v.windowHeight - (lipgloss.Height(v.headerView(v.windowWidth, ANNOTATIONS)) + lipgloss.Height(footerView(v.windowWidth, v.annotationsViewPort)))
	v.annotationsViewPort.Width = v.windowWidth

	v.labelsViewPort.Height = v.windowHeight - (lipgloss.Height(v.headerView(v.windowWidth, LABELS)) + lipgloss.Height(footerView(v.windowWidth, v.labelsViewPort)))
	v.labelsViewPort.Width = v.windowWidth

	v.eventsViewPort.Height = v.windowHeight - (lipgloss.Height(v.headerView(v.windowWidth, EVENTS)) + lipgloss.Height(footerView(v.windowWidth, v.eventsViewPort)))
	v.eventsViewPort.Width = v.windowWidth

	v.logsViewPort.Height = v.windowHeight - (lipgloss.Height(v.headerView(v.windowWidth, LOGS)) + lipgloss.Height(footerView(v.windowWidth, v.logsViewPort)))
	v.logsViewPort.Width = v.windowWidth

	if _, ok := v.pod.Logs[v.selectedContainer]; ok && v.logsViewPort.Height > 0 {
		v.logsViewPort.SetContent(strings.Join(jsoncolor.JSONLines(v.windowWidth, v.pod.Logs[v.selectedContainer]), "\n\n"))
		v.logsViewPort.GotoBottom()
	}

	if v.annotationsViewPort.Height > 0 {
		v.annotationsViewPort.SetContent(table.RowsToString(stringMapColumnsAndRows(v.windowWidth, "Key", "Value", v.pod.Pod.Annotations)))
	}

	if v.labelsViewPort.Height > 0 {
		v.labelsViewPort.SetContent(table.RowsToString(stringMapColumnsAndRows(v.windowWidth, "Key", "Value", v.pod.Pod.Labels)))
	}

	if v.eventsViewPort.Height > 0 {
		v.eventsViewPort.SetContent(table.RowsToString(eventColumnsAndRows(v.windowWidth, v.pod.Events)))
	}

	return v
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
	if v.tab > 0 {
		v.tab--
	} else {
		v.tab = tab(len(v.tabs) - 1)
	}

	return v
}

func (v View) moveTabRight() View {
	if v.tab < tab(len(v.tabs)-1) {
		v.tab++
	} else {
		v.tab = 0
	}

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
		return help.Full(v.windowWidth, v.fullHelp())
	}

	builder := strings.Builder{}
	header := v.headerView(v.windowWidth, v.tab)
	builder.WriteString(header)

	if v.pod == nil {
		return builder.String()
	}

	switch v.tab {
	case STATUS:
		columns, row := podStatusColumnsAndRows(v.pod.Pod)
		builder.WriteString(table.RowsToString(columns, []table.DataRow{row}))
		return builder.String()

	case ANNOTATIONS:
		footer := footerView(v.windowWidth, v.annotationsViewPort)
		builder.WriteString(v.annotationsViewPort.View())
		builder.WriteString(footer)

	case LABELS:
		footer := footerView(v.windowWidth, v.labelsViewPort)
		builder.WriteString(v.labelsViewPort.View())
		builder.WriteString(footer)

	case EVENTS:
		footer := footerView(v.windowWidth, v.labelsViewPort)
		builder.WriteString(v.eventsViewPort.View())
		builder.WriteString(footer)

	case LOGS:
		footer := footerView(v.windowWidth, v.logsViewPort)
		builder.WriteString(v.logsViewPort.View())
		builder.WriteString(footer)
	}

	return builder.String()
}

func (v View) headerView(width int, forTab tab) string {
	if v.pod == nil {
		return "Loading..."
	}

	builder := strings.Builder{}

	builder.WriteString(help.Short(width, []key.Binding{
		v.keys.Help,
		v.keys.Quit,
		v.keys.Refresh,
		v.keys.Left,
		v.keys.Right,
	}))

	builder.WriteString("\n\n")

	builder.WriteString(selection.Tabs(int(forTab), width, v.tabs) + "\n\n")

	if forTab == LOGS {
		builder.WriteString(selection.HorizontalList(v.containerNames, v.selectedContainer, width))
		builder.WriteString("\n")
	}

	builder.WriteString(tableHeaderView(width, forTab, *v.pod))

	return builder.String()
}

// tableHeaderView creates the table header view.
// Producing table headers seperately from the rows allows us to let the content scroll past the headers without hiding them.
func tableHeaderView(width int, t tab, pod pods.Pod) string {
	var columns []table.DataColumn
	switch t {
	case STATUS:
		columns, _ = podStatusColumnsAndRows(pod.Pod)
	case ANNOTATIONS:
		columns, _ = stringMapColumnsAndRows(width, "Key", "Value", pod.Pod.Annotations)
	case LABELS:
		columns, _ = stringMapColumnsAndRows(width, "Key", "Value", pod.Pod.Labels)
	case EVENTS:
		columns, _ = eventColumnsAndRows(width, pod.Events)
	case LOGS:
		return strings.Repeat("─", width) + "\n"
	}

	line := strings.Repeat("─", width)
	return lipgloss.NewStyle().Width(width).Render(table.ColumnsToString(columns)) + "\n" + lipgloss.JoinHorizontal(lipgloss.Center, line) + "\n\n"
}

// footerView creates the footerView which contains information about how far the user has scrolled through the viewPort.
func footerView(width int, viewPort viewport.Model) string {
	info := fmt.Sprintf("%3.f%%", viewPort.ScrollPercent()*100)
	line := strings.Repeat("─", integer.IntMax(0, width-lipgloss.Width(info)))
	return "\x1b[0m" + "\n" + lipgloss.NewStyle().Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Center, line, info))
}

// Init initializes the view.
func (v View) Init(c kubeui.Context) tea.Cmd {
	pod, err := v.k8sClient.GetPod(c.Namespace, c.SelectedPod)
	if err != nil {
		return kubeui.Error(err)
	}

	return kubeui.GenericCmd(k8smsg.NewGetPodMsg(pod))
}

// Destroy is called before a view is removed as the active view in the application.
func (v View) Destroy(c kubeui.Context) tea.Cmd {
	return nil
}
