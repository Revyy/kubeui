package pods

import (
	"fmt"
	"kubeui/internal/app/pods/views/errorinfo"
	"kubeui/internal/app/pods/views/namespaceselection"
	"kubeui/internal/app/pods/views/podinfo"
	"kubeui/internal/app/pods/views/podselection"
	"kubeui/internal/pkg/kubeui"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Model defines the base Model of the application.
type Model struct {
	kubeuiContext kubeui.Context

	// currentView is the currently displayed view.
	currentView kubeui.View

	initializing bool
	errorMessage string
}

// NewModel creates a new model.
func NewModel(rawConfig api.Config, configAccess clientcmd.ConfigAccess, clientSet *kubernetes.Clientset) *Model {
	return &Model{
		kubeuiContext: kubeui.Context{
			ConfigAccess: configAccess,
			Kubectl:      clientSet,
			Namespace:    "default",
			ApiConfig:    rawConfig,
		},
		initializing: true,
	}
}

// windowSizeUpdate handles updates to the terminal window size.
func (m Model) windowSizeUpdate(windowSize tea.WindowSizeMsg) Model {
	m.kubeuiContext.WindowWidth = windowSize.Width
	m.kubeuiContext.WindowHeight = windowSize.Height
	return m
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Global Keypresses and app messages.
	switch msg := msg.(type) {
	case Initialize:
		currentContext, ok := m.kubeuiContext.ApiConfig.Contexts[m.kubeuiContext.ApiConfig.CurrentContext]

		if !ok {
			return m, kubeui.Error(fmt.Errorf("invalid context"))
		}

		m.kubeuiContext.Namespace = currentContext.Namespace

		if m.kubeuiContext.Namespace == "default" {
			return m, kubeui.PushView("namespace_selection")
		}

		return m, kubeui.PushView("pod_selection")

	case tea.WindowSizeMsg:
		return m.windowSizeUpdate(msg), nil
	case error:
		m.errorMessage = msg.Error()
		return m, kubeui.PushView("error_info")

	case kubeui.PushViewMsg:
		if m.initializing {
			m.initializing = false
		}
		m.currentView = m.initializeView(msg.Id)
		return m, m.currentView.Init(m.kubeuiContext)
	}

	c, v, cmd := m.currentView.Update(m.kubeuiContext, kubeui.Msg{TeaMsg: msg})

	m.kubeuiContext = c
	m.currentView = v

	return m, cmd
}

func (m Model) initializeView(viewId string) kubeui.View {
	switch viewId {
	case "pod_selection":
		return podselection.New()
	case "namespace_selection":
		return namespaceselection.New()
	case "pod_info":
		return podinfo.New(m.kubeuiContext)
	case "error_info":
		return errorinfo.New(m.errorMessage)
	}

	return namespaceselection.New()
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (m Model) View() string {
	if m.initializing {
		return "Initializing..."
	}

	return m.currentView.View(m.kubeuiContext)
}

type Initialize struct{}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (m Model) Init() tea.Cmd {

	return func() tea.Msg {
		return Initialize{}
	}
	//return k8scommand.ListNamespaces(m.kubeuiContext.Kubectl)
}
