package podselection

import (
	"fmt"

	"kubeui/internal/pkg/component/columntable"
	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/k8scommand"
	"kubeui/internal/pkg/kubeui"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

// keyMap defines the keys that are handled by this view.
type keyMap struct {
	kubeui.GlobalKeyMap
	SelectNamespace key.Binding
}

// newKeyMap defines the actual key bindings and creates a keyMap.
func newKeyMap() *keyMap {
	return &keyMap{
		GlobalKeyMap: kubeui.NewGlobalKeyMap(),
		SelectNamespace: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "Select namespace"),
		),
	}
}

func (v View) fullHelp() [][]key.Binding {

	bindings := [][]key.Binding{
		{v.keys.Help, v.keys.Quit, v.keys.Refresh},
	}

	bindings[0] = append(bindings[0], v.keys.SelectNamespace)

	if len(v.pods) > 0 {
		bindings = append(bindings, v.podTable.KeyList())
	}

	return bindings
}

// New creates a new View.
func New() View {
	return View{
		keys:    newKeyMap(),
		loading: true,
	}
}

// View is used to select a pod.
type View struct {
	keys *keyMap

	// Pods in current namespace.
	pods []v1.Pod

	// Dialog used to confirm.
	activeDialog *confirm.Model

	// ColumnTable used to select a pod.
	podTable columntable.Model

	// Loading indicator
	initialized bool
	loading     bool

	// Show full help view or not.
	showFullHelp bool
}

// Update handles new messages from the runtime.
func (v View) Update(c kubeui.Context, msg kubeui.Msg) (kubeui.Context, kubeui.View, tea.Cmd) {

	if msg.IsKeyMsg() && v.showFullHelp {
		v.showFullHelp = false
		return c, v, nil
	}

	if msg.MatchesKeyBindings(v.keys.Help) && !v.showFullHelp {
		v.showFullHelp = true
		return c, v, nil
	}

	if msg.MatchesKeyBindings(v.keys.Quit) {
		return c, v, kubeui.Exit()
	}

	if msg.MatchesKeyBindings(v.keys.SelectNamespace) {
		return c, v, kubeui.PushView("namespace_selection")
	}

	if msg.MatchesKeyBindings(v.keys.Refresh) {
		return c, v, k8scommand.ListPods(c.Kubectl, c.Namespace)
	}

	// Results
	switch t := msg.TeaMsg.(type) {
	case k8scommand.ListPodsMsg:
		v.pods = t.PodList.Items
		podColumns, podRows := podTableContents(v.pods)
		var cmd tea.Cmd

		// The first time we receive a list of pods then we create a new podTable.
		// Otherwise we just update it.
		if !v.initialized {
			v.initialized = true
			v.podTable = columntable.New(podColumns, podRows, 10, "", true, columntable.Options{SingularItemName: "pod", StartInSearchMode: true})
		} else {
			v.podTable, cmd = v.podTable.Update(columntable.UpdateRowsAndColumns{Rows: podRows, Columns: podColumns})
		}

		v.loading = false

		return c, v, cmd

	case columntable.Selection:
		c.SelectedPod = t.Id
		return c, v, kubeui.PushView("pod_info")

	// When the user tries to delete a pod we create a new confirmation dialog and move to the CONFIRM_POD_DELETION state which will
	// display the dialog and handle the choice.
	case columntable.Deletion:
		dialog := confirm.New([]confirm.Button{{Desc: "Yes", Id: t.Id}, {Desc: "No", Id: t.Id}}, fmt.Sprintf("Are you sure you want to delete %s", t.Id))
		v.activeDialog = &dialog
		return c, v, nil

	// When a pod is actually deleted we refresh the pod list by returning the listPods command.
	case k8scommand.PodDeletedMsg:
		return c, v, k8scommand.ListPods(c.Kubectl, c.Namespace)

	case confirm.ButtonPress:
		v.activeDialog = nil
		if t.Pressed.Desc == "Yes" {
			return c, v, k8scommand.DeletePod(c.Kubectl, c.Namespace, t.Pressed.Id)
		}
		return c, v, nil
	}

	// If we have an active dialog.
	if v.activeDialog != nil {
		dialog, cmd := v.activeDialog.Update(msg.TeaMsg)
		v.activeDialog = &dialog
		return c, v, cmd
	}

	var cmd tea.Cmd
	v.podTable, cmd = v.podTable.Update(msg.TeaMsg)
	return c, v, cmd
}

// podTableContents creates the neccessary columns and rows for the columntable in order to display pod information.
func podTableContents(pods []v1.Pod) ([]*columntable.Column, []*columntable.Row) {
	podColumns := []*columntable.Column{
		{Desc: "Name", Width: 4},
		{Desc: "Ready", Width: 5},
		{Desc: "Status", Width: 6},
		{Desc: "Restarts", Width: 8},
		{Desc: "Age", Width: 3},
	}

	podRows := slices.Map(pods, func(p v1.Pod) *columntable.Row {

		podFormat := k8s.NewListPodFormat(p)

		// Update widths of the name and status columns
		podColumns[0].Width = integer.IntMax(podColumns[0].Width, len(p.Name))
		podColumns[1].Width = integer.IntMax(podColumns[1].Width, len(podFormat.Ready))
		podColumns[2].Width = integer.IntMax(podColumns[2].Width, len(podFormat.Status))
		podColumns[3].Width = integer.IntMax(podColumns[3].Width, len(podFormat.Restarts))
		podColumns[4].Width = integer.IntMax(podColumns[4].Width, len(podFormat.Age))

		return &columntable.Row{
			Id:     p.Name,
			Values: []string{podFormat.Name, podFormat.Ready, podFormat.Status, podFormat.Restarts, podFormat.Age},
		}
	})

	return podColumns, podRows
}

// View renders the ui of the view.
func (v View) View(c kubeui.Context) string {

	if v.showFullHelp {
		return kubeui.FullHelp(c.WindowWidth, v.fullHelp())
	}

	builder := strings.Builder{}

	builder.WriteString(kubeui.ShortHelp(c.WindowWidth, []key.Binding{v.keys.Help, v.keys.Quit, v.keys.SelectNamespace}))
	builder.WriteString("\n\n")

	if v.activeDialog != nil {
		builder.WriteString(v.activeDialog.View())
		return builder.String()
	}

	podViewStatusBar := kubeui.StatusBar(c.WindowWidth-1, " ", fmt.Sprintf("Context: %s  Namespace: %s", c.ApiConfig.CurrentContext, c.Namespace))
	builder.WriteString(podViewStatusBar + "\n")

	if v.loading {
		return "Loading..."
	} else if len(v.pods) == 0 {
		builder.WriteString(fmt.Sprintf("No pods found in namespace %s", c.Namespace))
	} else {
		builder.WriteString(v.podTable.View())
	}

	return builder.String()
}

// Init initializes the view.
func (v View) Init(c kubeui.Context) tea.Cmd {
	return k8scommand.ListPods(c.Kubectl, c.Namespace)
}

// Destroy is called before a view is removed as the active view in the application.
func (v View) Destroy(c kubeui.Context) tea.Cmd {
	return nil
}
