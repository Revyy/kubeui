package namespaceselection

import (
	"fmt"
	"strings"

	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8smsg"
	"kubeui/internal/pkg/kubeui"
	"kubeui/internal/pkg/ui/help"
	"kubeui/internal/pkg/ui/statusbar"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
)

// keyMap defines the keys that are handled by this view.
type keyMap struct {
	kubeui.GlobalKeyMap
}

// newKeyMap defines the actual key bindings and creates a keyMap.
func newKeyMap() *keyMap {
	return &keyMap{
		GlobalKeyMap: kubeui.NewGlobalKeyMap(),
	}
}

func (v View) fullHelp() [][]key.Binding {
	bindings := [][]key.Binding{
		{v.keys.Help, v.keys.Quit, v.keys.ExitView},
	}
	bindings = append(bindings, v.namespaceTable.KeyList())

	return bindings
}

// selectedNamespaceMsg is sent when a namespace has been selected and the k8s context has successfully been updated.
type selectedNamespaceMsg string

// K8sClient represents the interface towards kubernetes needed by this view.
type K8sClient interface {
	ListNamespaces() (*v1.NamespaceList, error)
}

// ContextClient represents the interface for working with kubernetes contexts that the view needs.
type ContextClient interface {
	CurrentContext() string
	SwitchContext(ctx, namespace string) (err error)
}

// View allows the user to select a namespace.
type View struct {
	windowHeight int
	windowWidth  int

	keys *keyMap

	// Namespaces in current cluster.
	namespaces []string

	// SearchTable used to select a namespace.
	namespaceTable searchtable.Model

	// Show full help view or not.
	showFullHelp bool

	// If the View has been initialized or not.
	initialized bool

	// Kubernetes client.
	k8sClient K8sClient

	// KubeContext client.
	contextClient ContextClient
}

// New creates a new View.
func New(k8sClient K8sClient, contextClient ContextClient, windowWidth, windowHeight int) View {
	return View{
		k8sClient:     k8sClient,
		contextClient: contextClient,
		windowWidth:   windowWidth,
		windowHeight:  windowHeight,
		keys:          newKeyMap(),
	}
}

// Update handles new messages from the runtime.
func (v View) Update(c kubeui.Context, msg kubeui.Msg) (kubeui.Context, kubeui.View, tea.Cmd) {
	if msg.IsWindowResize() {
		windowResizeMsg, ok := msg.GetWindowResizeMsg()

		if !ok {
			return c, v, nil
		}

		v.windowHeight = windowResizeMsg.Height
		v.windowWidth = windowResizeMsg.Width

		return c, v, nil
	}

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

	if msg.MatchesKeyBindings(v.keys.ExitView) {
		// We don't reinitialize the pod selection view when exiting the view.
		return c, v, kubeui.PushView("pod_selection", false)
	}

	// Results
	switch t := msg.TeaMsg.(type) {

	case k8smsg.ListNamespacesMsg:
		v.namespaces = slices.Map(t.NamespaceList.Items, func(n v1.Namespace) string {
			return n.GetName()
		})
		v.namespaceTable = searchtable.New(
			v.namespaces,
			10,
			c.Namespace,
			false,
			searchtable.Options{
				SingularItemName:  "namespace",
				StartInSearchMode: true,
			},
		)
		v.initialized = true
		return c, v, nil

	case searchtable.Selection:
		return c, v, func() tea.Msg {
			err := v.contextClient.SwitchContext(v.contextClient.CurrentContext(), t.Value)
			if err != nil {
				return err
			}
			return selectedNamespaceMsg(t.Value)
		}

	case selectedNamespaceMsg:
		c.Namespace = string(t)
		// If we have made a selection then we reinitialize the pod selection view to load the pods for that namespace.
		return c, v, kubeui.PushView("pod_selection", true)

	}

	var cmd tea.Cmd
	v.namespaceTable, cmd = v.namespaceTable.Update(msg.TeaMsg)
	return c, v, cmd
}

// View renders the ui of the view.
func (v View) View(c kubeui.Context) string {
	if v.showFullHelp {
		return help.Full(v.windowWidth, v.fullHelp())
	}

	builder := strings.Builder{}

	builder.WriteString(help.Short(v.windowWidth, []key.Binding{v.keys.Help, v.keys.Quit, v.keys.ExitView}))
	builder.WriteString("\n\n")

	statusBar := statusbar.New(v.windowWidth-1, " ", fmt.Sprintf("Context: %s", v.contextClient.CurrentContext()))
	builder.WriteString(statusBar + "\n")

	builder.WriteString(v.namespaceTable.View())

	return builder.String()
}

// Init initializes the view.
func (v View) Init(c kubeui.Context) tea.Cmd {
	if v.initialized {
		return nil
	}

	return func() tea.Msg {
		namespaces, err := v.k8sClient.ListNamespaces()
		if err != nil {
			return err
		}

		return k8smsg.NewListNamespacesMsg(namespaces)
	}
}

// Destroy is called before a view is removed as the active view in the application.
func (v View) Destroy(c kubeui.Context) tea.Cmd {
	return nil
}
