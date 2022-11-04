package pods

import (
	"fmt"
	"strings"

	"kubeui/internal/pkg/component/columntable"
	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/component/podview"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/k8scommand"
	"kubeui/internal/pkg/kubeui"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// AppState defines the different states the application can be in.
type AppState uint16

const (
	// INITIALIZING represents the initial state of the application where initial loading of data occurs.
	// However it is also used to reinitialize the application when a new namespace is selected.
	INITIALIZING AppState = iota
	// When the application is in the POD_SELECTION state it shows a table with information about all pods in the current namespace.
	POD_SELECTION
	// When the application is in the NAMESPACE_SELECTION state it allows the user to select a namespace.
	NAMESPACE_SELECTION
	// When the application is in the NAMESPACE_SELECTION state it allows the user to confirm or deny a pod deletion request.
	CONFIRM_POD_DELETION
	// When a pod has been selected and is being viewed.
	PODVIEW
	// When displaying the full help.
	FULLHELP
	// When the application is in the ERROR state it allows the user to view an error message before quitting the application.
	ERROR
)

// Model defines the base Model of the application.
type Model struct {
	// application level keybindings
	keys *appKeyMap

	// kubernetes config object.
	config api.Config

	// object defining how the kubernetes config was located and put together.
	// needed in order to modify the config files on disc.
	configAccess clientcmd.ConfigAccess

	// ClientSet used to issue commands to kubernetes.
	kubectl *kubernetes.Clientset

	// Windows size.
	windowSize tea.WindowSizeMsg

	// SearchTable used to select a namespace.
	namespaceTable searchtable.Model

	// ColumnTable used to select a pod.
	podTable columntable.Model

	// PodView used to visualize a pod.
	podView podview.Model

	// The currently selected pod if any.
	currentPod k8s.Pod

	// Dialog used to confirm.
	activeDialog *confirm.Model

	// Indicates which state the application is in.
	state AppState

	// The previous state of the application.
	prevState AppState

	// Error message to be displayed.
	errorMessage string

	// Namespaces in current cluster.
	namespaces []string

	// Pods in current namespace.
	pods []v1.Pod

	// Currently selected namespace.
	currentNamespace string
}

// NewModel creates a new model.
func NewModel(rawConfig api.Config, configAccess clientcmd.ConfigAccess, clientSet *kubernetes.Clientset) *Model {
	return &Model{
		keys:             newAppKeyMap(),
		config:           rawConfig,
		configAccess:     configAccess,
		kubectl:          clientSet,
		currentNamespace: "default",
		state:            INITIALIZING,
	}
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (m Model) ShortHelp() []key.Binding {

	if m.state == ERROR {
		return []key.Binding{m.keys.Quit}
	}

	bindings := []key.Binding{m.keys.Help, m.keys.Quit}

	if m.state == POD_SELECTION {
		bindings = append(bindings, m.keys.refreshPodList, m.keys.selectNamespace)
	}

	if m.state == PODVIEW {
		bindings = append(bindings, m.keys.exitView)
	}

	return bindings
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (m Model) FullHelp() [][]key.Binding {

	// We look at prevState here as we are in the FULLHELP state and want to display the full help of the previous state
	// before going back there.
	if m.prevState == ERROR {
		return [][]key.Binding{{m.keys.Quit}}
	}

	bindings := [][]key.Binding{
		{m.keys.Help, m.keys.Quit},
	}

	switch m.prevState {
	case NAMESPACE_SELECTION:
		bindings = append(bindings, m.namespaceTable.KeyList())
	case POD_SELECTION:

		bindings[0] = append(bindings[0], m.keys.selectNamespace, m.keys.refreshPodList)

		if len(m.pods) > 0 {
			bindings = append(bindings, m.podTable.KeyList())
		}

	case CONFIRM_POD_DELETION:
		if m.activeDialog != nil {
			bindings = append(bindings, m.activeDialog.KeyList())
		}
	case PODVIEW:
		bindings[0] = append(bindings[0], m.keys.exitView)
		bindings = append(bindings, m.podView.KeyList())
	}
	return bindings
}

// updateState is used to set a new state.
func (m Model) updateState(newState AppState) Model {
	m.prevState = m.state
	m.state = newState

	return m
}

// windowSizeUpdate handles updates to the terminal window size.
func (m Model) windowSizeUpdate(windowSize tea.WindowSizeMsg) Model {
	m.windowSize = windowSize
	return m
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// If we are in the error state we only allow quitting.
	if m.state == ERROR {
		if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, m.keys.Quit) {
			return m, tea.Quit
		}
		return m, nil
	}

	// Global Keypresses and app messages.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// The podview requires access to window resize message in order to adjust the size of its viewport.
		m.podView, _ = m.podView.Update(msg)
		return m.windowSizeUpdate(msg), nil
	case error:
		m = m.updateState(ERROR)
		m.errorMessage = msg.Error()
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			if m.state == FULLHELP {
				m = m.updateState(m.prevState)
			} else {
				m = m.updateState(FULLHELP)
			}
		// We can only transition to NAMESPACE_SELECTION from POD_SELECTION.
		case key.Matches(msg, m.keys.selectNamespace) && m.state == POD_SELECTION:
			m.namespaceTable = searchtable.New(
				m.namespaces,
				10,
				m.currentNamespace,
				false,
				searchtable.Options{
					SingularItemName:  "namespace",
					StartInSearchMode: true,
				},
			)
			m = m.updateState(NAMESPACE_SELECTION)
			return m, nil
		}
	}

	// State specific updates.
	var cmd tea.Cmd

	switch m.state {
	case INITIALIZING:
		m, cmd = m.initializingUpdate(msg)
	case NAMESPACE_SELECTION:
		m, cmd = m.namespaceSelectionUpdate(msg)
	case POD_SELECTION:
		m, cmd = m.podSelectionUpdate(msg)
	case CONFIRM_POD_DELETION:
		m, cmd = m.confirmPodDeletionUpdate(msg)
	case PODVIEW:
		m, cmd = m.podViewUpdate(msg)
	}

	return m, cmd
}

// initializingUpdate handles updates for the INITIALIZING app state.
func (m Model) initializingUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case k8scommand.ListNamespacesMsg:
		m.namespaces = slices.Map(msgT.NamespaceList.Items, func(n v1.Namespace) string {
			return n.GetName()
		})

		currentContext, ok := m.config.Contexts[m.config.CurrentContext]

		if ok && slices.Contains(m.namespaces, currentContext.Namespace) {
			m.currentNamespace = currentContext.Namespace
		}

		return m, k8scommand.ListPods(m.kubectl, m.currentNamespace)
	// This is the result message of m.listPods.
	case k8scommand.ListPodsMsg:
		m.pods = msgT.PodList.Items
		podColumns, podRows := podTableContents(m.pods)
		m.podTable = columntable.New(podColumns, podRows, 10, "", true, columntable.Options{SingularItemName: "pod", StartInSearchMode: true})
		m = m.updateState(POD_SELECTION)

		return m, nil

	}

	return m, nil
}

// namespaceSelectionUpdate handles updates for the NAMESPACE_SELECTION app state.
func (m Model) namespaceSelectionUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case searchtable.Selection:

		err := k8s.SwitchContext(m.config.CurrentContext, msgT.Value, m.configAccess, m.config)
		if err != nil {
			return m, func() tea.Msg { return err }
		}

		m.currentNamespace = msgT.Value
		m = m.updateState(POD_SELECTION)
		return m, k8scommand.ListPods(m.kubectl, m.currentNamespace)
	}

	var cmd tea.Cmd
	m.namespaceTable, cmd = m.namespaceTable.Update(msg)
	return m, cmd
}

// podSelectionUpdate handles updates for the POD_SELECTION app state.
func (m Model) podSelectionUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	// The message sent after a new list of pods have been fetched using the listPods command.
	case k8scommand.ListPodsMsg:
		m.pods = msgT.PodList.Items
		podColumns, podRows := podTableContents(m.pods)

		var cmd tea.Cmd
		m.podTable, cmd = m.podTable.Update(columntable.UpdateRowsAndColumns{Rows: podRows, Columns: podColumns})
		return m, cmd

	// When a pod is selected we fetch an updated version of the pod.
	case columntable.Selection:
		return m, k8scommand.GetPod(m.kubectl, m.currentNamespace, msgT.Id)

	// When the pod is fetched then we move the state to PODVIEW and create a new podview to view the state of the pod.
	case k8scommand.GetPodMsg:
		m = m.updateState(PODVIEW)
		m.podView = podview.New(*msgT.Pod, lipgloss.Height(m.headerView()), m.windowSize.Width, m.windowSize.Height)
		m.currentPod = *msgT.Pod
		return m, m.podView.Init()

	// When the user tries to delete a pod we create a new confirmation dialog and move to the CONFIRM_POD_DELETION state which will
	// display the dialog and handle the choice.
	case columntable.Deletion:
		dialog := confirm.New([]confirm.Button{{Desc: "Yes", Id: msgT.Id}, {Desc: "No", Id: msgT.Id}}, fmt.Sprintf("Are you sure you want to delete %s", msgT.Id))
		m.activeDialog = &dialog
		m = m.updateState(CONFIRM_POD_DELETION)
		return m, nil

	// When a pod is actually deleted we refresh the pod list by returning the listPods command.
	case k8scommand.PodDeletedMsg:
		return m, k8scommand.ListPods(m.kubectl, m.currentNamespace)

	// Refresh the pod list.
	case tea.KeyMsg:
		if key.Matches(msgT, m.keys.refreshPodList) {
			return m, k8scommand.ListPods(m.kubectl, m.currentNamespace)
		}
	}

	var cmd tea.Cmd
	m.podTable, cmd = m.podTable.Update(msg)
	return m, cmd
}

// podViewUpdate handles updates for the PODVIEW app state.
func (m Model) podViewUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msgT, m.keys.exitView) {
			m = m.updateState(POD_SELECTION)
			return m, k8scommand.ListPods(m.kubectl, m.currentNamespace)
		}
	case podview.Refresh:
		return m, k8scommand.GetPod(m.kubectl, m.currentNamespace, msgT.PodName)
	case k8scommand.GetPodMsg:
		var cmd tea.Cmd
		m.podView, cmd = m.podView.Update(podview.NewPod{Pod: *msgT.Pod})
		m.currentPod = *msgT.Pod
		return m, cmd
	}

	var cmd tea.Cmd
	m.podView, cmd = m.podView.Update(msg)
	return m, cmd
}

// confirmPodDeletionUpdate handles updates for the CONFIRM_POD_DELETION app state.
func (m Model) confirmPodDeletionUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case confirm.ButtonPress:
		m = m.updateState(POD_SELECTION)
		m.activeDialog = nil
		if msgT.Pressed.Desc == "Yes" {
			return m, k8scommand.DeletePod(m.kubectl, m.currentNamespace, msgT.Pressed.Id)
		}

		return m, nil
	}

	dialog, cmd := m.activeDialog.Update(msg)
	m.activeDialog = &dialog

	return m, cmd
}

// headerView builds a view containing basic help information and a status bar for some states.
func (m Model) headerView() string {
	builder := strings.Builder{}

	builder.WriteString(kubeui.ShortHelp(m.windowSize.Width, m.ShortHelp()))
	builder.WriteString("\n\n")

	if slices.Contains([]AppState{CONFIRM_POD_DELETION, INITIALIZING, ERROR}, m.state) {
		return builder.String()
	}

	switch m.state {
	case PODVIEW:
		podViewStatusBar := kubeui.StatusBar(m.windowSize.Width-1, " ", fmt.Sprintf("Context: %s  Namespace: %s Pod: %s", m.config.CurrentContext, m.currentNamespace, m.currentPod.Pod.GetName()))
		builder.WriteString(podViewStatusBar + "\n")
	default:
		baseStatusBar := kubeui.StatusBar(m.windowSize.Width-1, " ", fmt.Sprintf("Context: %s  Namespace: %s", m.config.CurrentContext, m.currentNamespace))
		builder.WriteString(baseStatusBar + "\n")
	}

	return builder.String()
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (m Model) View() string {

	builder := strings.Builder{}

	switch m.state {
	case INITIALIZING:
		builder.WriteString("Loading...")
	case ERROR:
		builder.WriteString(m.headerView())
		builder.WriteString("An error occured\n\n")
		builder.WriteString(kubeui.ErrorMessageStyle.Render(kubeui.LineBreak(m.errorMessage, m.windowSize.Width)))

	case CONFIRM_POD_DELETION:
		builder.WriteString(m.headerView())
		if m.activeDialog != nil {
			builder.WriteString(m.activeDialog.View())
		}
	case POD_SELECTION:
		builder.WriteString(m.headerView())
		if len(m.pods) == 0 {
			builder.WriteString(fmt.Sprintf("No pods found in namespace %s", m.currentNamespace))
			break
		}
		builder.WriteString(m.podTable.View())
	case NAMESPACE_SELECTION:
		builder.WriteString(m.headerView())
		builder.WriteString(m.namespaceTable.View())
	case PODVIEW:
		builder.WriteString(m.headerView())
		builder.WriteString(m.podView.View())
	case FULLHELP:
		builder.WriteString(kubeui.FullHelp(m.windowSize.Width, m.FullHelp()))
	}

	return builder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (m Model) Init() tea.Cmd {
	return k8scommand.ListNamespaces(m.kubectl)
}
