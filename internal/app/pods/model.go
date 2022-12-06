package pods

import (
	"fmt"
	"kubeui/internal/app/pods/views/errorinfo"
	"kubeui/internal/app/pods/views/namespaceselection"
	"kubeui/internal/app/pods/views/podinfo"
	"kubeui/internal/app/pods/views/podselection"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/kubeui"

	tea "github.com/charmbracelet/bubbletea"
)

// Model defines the base Model of the application.
type Model struct {
	windowHeight int
	windowWidth  int

	kubeuiContext kubeui.Context

	// currentView is the currently displayed view.
	currentView string
	// previousView is the previously displayed view.
	previousView string

	initializing bool
	errorMessage string

	contextClient k8s.ContextClient
	k8sClient     k8s.Service

	views map[string]kubeui.View
}

// NewModel creates a new model.
func NewModel(contextClient k8s.ContextClient, k8sClient k8s.Service) *Model {
	return &Model{
		kubeuiContext: kubeui.Context{
			Namespace: "default",
		},
		contextClient: contextClient,
		k8sClient:     k8sClient,
		views:         map[string]kubeui.View{},
		initializing:  true,
	}
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Global Keypresses and app messages.
	switch msgT := msg.(type) {
	case Initialize:
		currentContext, ok := m.contextClient.CurrentApiContext()

		if !ok {
			return m, kubeui.Error(fmt.Errorf("invalid context"))
		}

		m.kubeuiContext.Namespace = currentContext.Namespace

		if m.kubeuiContext.Namespace == "default" {
			return m, kubeui.PushView("namespace_selection", true)
		}

		return m, kubeui.PushView("pod_selection", true)

	case tea.WindowSizeMsg:

		m.windowHeight = msgT.Height
		m.windowWidth = msgT.Width

		for k, v := range m.views {
			_, v, _ := v.Update(m.kubeuiContext, kubeui.Msg{TeaMsg: msg})
			m.views[k] = v
		}
		return m, nil
	case error:
		m.errorMessage = msgT.Error()
		return m, kubeui.PushView("error_info", true)

	case kubeui.PushViewMsg:
		if m.initializing {
			m.initializing = false
		}

		_, ok := m.views[msgT.Id]

		if !ok || msgT.Initialize {
			m.views[msgT.Id] = m.initializeView(msgT.Id)
		}

		// If this is the first view that was pushed then we set the previous view to the same as the new current view.
		m.previousView = m.currentView
		if m.previousView == "" {
			m.previousView = msgT.Id
		}

		m.currentView = msgT.Id

		if msgT.Initialize {
			return m, m.views[msgT.Id].Init(m.kubeuiContext)
		}

		return m, nil

	case kubeui.PopViewMsg:

		_, ok := m.views[m.previousView]

		if !ok {
			return m, kubeui.Error(fmt.Errorf("program error, invalid view"))
		}

		m.currentView = m.previousView
		m.previousView = ""

		if msgT.Initialize {
			m.views[m.currentView] = m.initializeView(m.currentView)
			return m, m.views[m.currentView].Init(m.kubeuiContext)
		}

		return m, nil
	}

	c, v, cmd := m.views[m.currentView].Update(m.kubeuiContext, kubeui.Msg{TeaMsg: msg})

	m.kubeuiContext = c
	m.views[m.currentView] = v

	return m, cmd
}

func (m Model) initializeView(viewId string) kubeui.View {
	switch viewId {
	case "pod_selection":
		return podselection.New(m.k8sClient, m.contextClient, m.windowWidth, m.windowHeight)
	case "namespace_selection":
		return namespaceselection.New(m.k8sClient, m.contextClient, m.windowWidth, m.windowHeight)
	case "pod_info":
		return podinfo.New(m.k8sClient, m.windowWidth, m.windowHeight)
	case "error_info":
		return errorinfo.New(m.errorMessage, m.windowWidth, m.windowHeight)
	}

	return namespaceselection.New(m.k8sClient, m.contextClient, m.windowWidth, m.windowHeight)
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (m Model) View() string {
	if m.initializing {
		return "Initializing..."
	}

	return m.views[m.currentView].View(m.kubeuiContext)
}

type Initialize struct{}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (m Model) Init() tea.Cmd {

	return func() tea.Msg {
		return Initialize{}
	}
}
