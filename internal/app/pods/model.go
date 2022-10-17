package pods

import (
	"fmt"
	"strings"

	"kubeui/internal/app/pods/message"
	"kubeui/internal/pkg/component/columntable"
	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/kubeui"
	"kubeui/internal/pkg/kubeui/help"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// appKeyMap defines the keys that are handled at the top level in the application.
// These keys will be checked before passing along a msg to underlying components.
type appKeyMap struct {
	quit            key.Binding
	help            key.Binding
	selectNamespace key.Binding
	refreshPodList  key.Binding
}

// newAppKeyMap defines the actual key bindings and creates an appKeyMap.
func newAppKeyMap() *appKeyMap {
	return &appKeyMap{
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c,ctrl+q", "Quit"),
		),
		help: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "Toggle help"),
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

	// Help.
	help help.Model

	// SearchTable used to select a namespace.
	namespaceTable searchtable.Model

	// ColumnTable used to select a pod.
	podTable columntable.Model

	// Dialog used to confirm.
	activeDialog *confirm.Model

	// Indicates which state the application is in.
	state AppState

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
		help:             help.New(),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (m Model) ShortHelp() []key.Binding {

	if m.state == ERROR {
		return []key.Binding{m.keys.quit}
	}

	bindings := []key.Binding{m.keys.help, m.keys.quit}

	if m.state == POD_SELECTION {
		bindings = append(bindings, m.keys.refreshPodList, m.keys.selectNamespace)
	}

	return bindings
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (m Model) FullHelp() [][]key.Binding {

	if m.state == ERROR {
		return [][]key.Binding{{m.keys.quit}}
	}

	bindings := [][]key.Binding{
		{m.keys.help, m.keys.quit},
	}

	switch m.state {
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
	}
	return bindings
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// If we are in the error state we only allow quitting.
	if m.state == ERROR {
		if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, m.keys.quit) {
			return m, tea.Quit
		}
		return m, nil
	}

	// Global Keypresses and app messages.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.windowSize = msg
		return m, nil
	case error:
		m.state = ERROR
		m.errorMessage = msg.Error()
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.help):
			m.help.ShowAll = !m.help.ShowAll
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
			m.state = NAMESPACE_SELECTION
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
	}

	return m, cmd
}

// initializingUpdate handles updates for the INITIALIZING app state.
func (m Model) initializingUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case message.Initialization:
		m.namespaces = slices.Map(msgT.NamespaceList.Items, func(n v1.Namespace) string {
			return n.GetName()
		})

		currentContext, ok := m.config.Contexts[m.config.CurrentContext]

		if ok && slices.Contains(m.namespaces, currentContext.Namespace) {
			m.currentNamespace = currentContext.Namespace
		}

		return m, m.listPods
	// This is the result message of m.listPods.
	case message.ListPods:
		m.pods = msgT.PodList.Items
		podColumns, podRows := podTableContents(m.pods)
		m.podTable = columntable.New(podColumns, podRows, 10, "", true, columntable.Options{SingularItemName: "pod", StartInSearchMode: true})
		m.state = POD_SELECTION

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
		m.state = POD_SELECTION
		return m, m.listPods
	}

	var cmd tea.Cmd
	m.namespaceTable, cmd = m.namespaceTable.Update(msg)
	return m, cmd
}

// podSelectionUpdate handles updates for the POD_SELECTION app state.
func (m Model) podSelectionUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case message.ListPods:
		m.pods = msgT.PodList.Items
		podColumns, podRows := podTableContents(m.pods)

		var cmd tea.Cmd
		m.podTable, cmd = m.podTable.Update(columntable.UpdateRowsAndColumns{Rows: podRows, Columns: podColumns})
		return m, cmd

	case columntable.Selection:
		m.state = ERROR
		m.errorMessage = fmt.Sprintf("you selected pod: %s", msgT.Id)
		return m, nil
	case columntable.Deletion:
		dialog := confirm.New([]confirm.Button{{Desc: "Yes", Id: msgT.Id}, {Desc: "No", Id: msgT.Id}}, fmt.Sprintf("Are you sure you want to delete %s", msgT.Id))
		m.activeDialog = &dialog
		m.state = CONFIRM_POD_DELETION
		return m, nil

	case message.PodDeleted:
		return m, m.listPods

	case tea.KeyMsg:
		if key.Matches(msgT, m.keys.refreshPodList) {
			return m, m.listPods
		}
	}

	var cmd tea.Cmd
	m.podTable, cmd = m.podTable.Update(msg)
	return m, cmd
}

// confirmPodDeletionUpdate handles updates for the CONFIRM_POD_DELETION app state.
func (m Model) confirmPodDeletionUpdate(msg tea.Msg) (Model, tea.Cmd) {

	switch msgT := msg.(type) {
	case confirm.ButtonPress:
		m.state = POD_SELECTION
		m.activeDialog = nil
		if msgT.Pressed.Desc == "Yes" {
			return m, m.deletePod(msgT.Pressed.Id)
		}

		return m, nil
	}

	dialog, cmd := m.activeDialog.Update(msg)
	m.activeDialog = &dialog

	return m, cmd
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (m Model) View() string {

	builder := strings.Builder{}

	helpView := m.help.View(m)
	builder.WriteString(helpView)
	builder.WriteString("\n\n")

	if m.state == CONFIRM_POD_DELETION && m.activeDialog != nil {
		builder.WriteString(m.activeDialog.View())
		return builder.String()
	}

	if m.state == INITIALIZING {
		builder.WriteString("Loading...")
		return builder.String()
	}

	if m.state == ERROR {
		builder.WriteString("An error occured\n\n")
		builder.WriteString(kubeui.ErrorMessageStyle.Render(kubeui.LineBreak(m.errorMessage, m.windowSize.Width)))
		return builder.String()
	}

	statusBar := kubeui.StatusBar(m.windowSize.Width-1, " ", fmt.Sprintf("Context: %s  Namespace: %s", m.config.CurrentContext, m.currentNamespace))
	builder.WriteString(statusBar + "\n")

	switch m.state {
	case POD_SELECTION:
		if len(m.pods) == 0 {
			builder.WriteString(fmt.Sprintf("No pods found in namespace %s", m.currentNamespace))
			break
		}
		builder.WriteString(m.podTable.View())

	case NAMESPACE_SELECTION:
		builder.WriteString(m.namespaceTable.View())

	}

	return builder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (m Model) Init() tea.Cmd {
	return m.listNamespaces
}
