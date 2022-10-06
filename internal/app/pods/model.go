package pods

import (
	"fmt"
	"strings"

	"kubeui/internal/app/pods/message"
	"kubeui/internal/pkg/component/searchtable"
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
	}
}

type AppState uint16

const (
	INITIALIZING AppState = iota
	MAIN
	NAMESPACE_SELECT
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

	// searchtable used to select a namespace.
	namespaceTable searchtable.SearchTable

	// searchtable used to select a pod.
	podTable searchtable.SearchTable

	// Indicates which state the application is in.
	state AppState

	// Error message to be displayed.
	errorMessage string

	// Namespaces in current cluster.
	namespaces []string

	// Pods in current namespace.
	pods []string

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

	return []key.Binding{m.keys.help, m.keys.quit, m.keys.selectNamespace}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (m Model) FullHelp() [][]key.Binding {

	if m.state == ERROR {
		return [][]key.Binding{{m.keys.quit}}
	}

	bindings := [][]key.Binding{
		{m.keys.help, m.keys.quit, m.keys.selectNamespace},
	}

	switch m.state {
	case NAMESPACE_SELECT:
		bindings = append(bindings, m.namespaceTable.KeyList())
	case MAIN:
		if len(m.pods) > 0 {
			bindings = append(bindings, m.podTable.KeyList())
		}
	}
	return bindings
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// If we are in the error state we only allow quitting.
	if m.state == ERROR {
		if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, m.keys.quit) {
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.help):
			//m.help.ShowAll = !m.help.ShowAll
			return m, func() tea.Msg {
				return fmt.Errorf("some error bla bla")
			}

		case key.Matches(msg, m.keys.selectNamespace):
			m.namespaceTable = searchtable.New(m.namespaces, 10, m.currentNamespace, false, searchtable.Options{SingularItemName: "namespace"})
			if m.state == NAMESPACE_SELECT {
				m.state = MAIN
			} else {
				m.state = NAMESPACE_SELECT
			}

			return m, nil
		}

	case error:
		m.state = ERROR
		m.errorMessage = msg.Error()
		return m, nil

	case message.Initialization:
		m.namespaces = slices.Map(msg.NamespaceList.Items, func(n v1.Namespace) string {
			return n.GetName()
		})

		currentContext, ok := m.config.Contexts[m.config.CurrentContext]

		if ok && slices.Contains(m.namespaces, currentContext.Namespace) {
			m.currentNamespace = currentContext.Namespace
		}

		return m, m.listPods

	case message.ListPods:
		m.pods = slices.Map(msg.PodList.Items, func(n v1.Pod) string {
			return n.GetName()
		})

		m.podTable = searchtable.New(m.pods, 10, "", false, searchtable.Options{SingularItemName: "pod"})

		if m.state != MAIN {
			m.state = MAIN
		}

		return m, nil
	}

	switch m.state {
	case NAMESPACE_SELECT:
		return m.namespaceSelectUpdate(msg)
	case MAIN:
		return m.podSelectUpdate(msg)
	}

	return m, nil

}

func (m Model) namespaceSelectUpdate(msg tea.Msg) (Model, tea.Cmd) {
	switch msgT := msg.(type) {
	case searchtable.Selection:
		m.currentNamespace = msgT.Value
		m.state = INITIALIZING
		return m, m.listPods
	}

	var cmd tea.Cmd
	m.namespaceTable, cmd = m.namespaceTable.Update(msg)
	return m, cmd
}

func (m Model) podSelectUpdate(msg tea.Msg) (Model, tea.Cmd) {
	/*switch msgT := msg.(type) {
	case searchtable.Selection:
		m.currentNamespace = msgT.Value
		m.state = MAIN
		return m, nil
	}*/

	var cmd tea.Cmd
	m.podTable, cmd = m.podTable.Update(msg)
	return m, cmd
}

func (m Model) View() string {

	builder := strings.Builder{}

	helpView := m.help.View(m)
	builder.WriteString(helpView)
	builder.WriteString("\n\n")

	if m.state == INITIALIZING {
		builder.WriteString("Loading...")
		return builder.String()
	}

	if m.state == ERROR {
		builder.WriteString("An error occured\n\n")
		builder.WriteString(errorMessageStyle.Render(kubeui.LineBreak(m.errorMessage, m.windowSize.Width)))
		return builder.String()
	}

	statusBar := statusBarStyle.Width(m.windowSize.Width - 1).Render(fmt.Sprintf("Namespace: %s", m.currentNamespace))
	builder.WriteString(statusBar + "\n")

	switch m.state {
	case MAIN:
		if len(m.pods) == 0 {
			builder.WriteString(fmt.Sprintf("No pods found in namespace %s", m.currentNamespace))
			break
		}
		builder.WriteString(m.podTable.View())

	case NAMESPACE_SELECT:
		builder.WriteString(m.namespaceTable.View())

	}

	return builder.String()
}

func (m Model) Init() tea.Cmd {
	return m.listNamespaces
}
