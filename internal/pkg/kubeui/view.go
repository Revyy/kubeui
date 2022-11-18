package kubeui

import (
	"kubeui/internal/pkg/k8s"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

// View is intended to be a stateful component that completely
// takes over the ui and handles most inputs except for some global keypresses and system messages.
type View interface {
	Init(Context) tea.Cmd
	Update(Context, Msg) (Context, View, tea.Cmd)
	View(Context) string
	Destroy(Context) tea.Cmd
}

// K8sClient defines the interface to fetch data from kubernetes.
type K8sClient interface {
	// Lists namespaces in the cluster.
	ListNamespaces() (*v1.NamespaceList, error)
	// Lists pods in the specified namespace.
	ListPods(namespace string) (*v1.PodList, error)
	// Fetches information about a single pod, including events and logs.
	GetPod(namespace, id string) (*k8s.Pod, error)
	// Delete the pod with the specified name in the specified namespace.
	// Returns the name of the deleted pod.
	DeletePod(namespace, name string) (string, error)
}

// ContextClient defines the interface for working with kubernetes contexts from a view.
type ContextClient interface {
	// Returns a list of available contexts.
	Contexts() []string
	// Returns the api.Context for the currently selected context if it exists.
	// If no api.Context exists for the current context then the bool should be set to false.
	CurrentApiContext() (*api.Context, bool)
	// Returns the currently selected context.
	CurrentContext() string
	// Switch to the specified context and optionally set the default namespace.
	SwitchContext(ctx, namespace string) (err error)
	// Delete the specified context.
	DeleteContext(ctx string) (err error)
	// Delete the specified user entry.
	DeleteUser(user string) (err error)
	// Delete the specified cluster entry.
	DeleteClusterEntry(cluster string) (err error)
}

// Context contains the context of the kubeui application.
type Context struct {
	// Used to read and manipulate the kubeconfig file and related contexts.
	ContextClient ContextClient

	// Used to issue commands and fetch data from kubernetes.
	K8sClient K8sClient

	// Currently selected namespace
	Namespace string

	// Name of currently selected pod.
	SelectedPod string
}

// Msg wraps the bubbletea message in a way that allows us to simplify some things.
type Msg struct {
	TeaMsg tea.Msg
}

// IsError tries to extract an error.
func (m Msg) IsError() (error, bool) {
	if e, ok := m.TeaMsg.(error); ok {
		return e, ok
	}
	return nil, false
}

// IsWindowResize checks if the msg contains a tea.WindowResizeMsg.
func (m Msg) IsWindowResize() bool {
	_, ok := m.TeaMsg.(tea.WindowSizeMsg)
	return ok
}

// GetWindowResizeMsg tries to extract a tea.WindowSizeMsg from the msg.
func (m Msg) GetWindowResizeMsg() (tea.WindowSizeMsg, bool) {
	w, ok := m.TeaMsg.(tea.WindowSizeMsg)
	return w, ok
}

// IsKeyMsg checks if the message contains a key click.
func (m Msg) IsKeyMsg() bool {

	_, ok := m.TeaMsg.(tea.KeyMsg)

	return ok
}

// MatchesKeyBindings checks if the message matches a specific KeyBinding.
func (m Msg) MatchesKeyBindings(bindings ...key.Binding) bool {

	keyMsg, ok := m.TeaMsg.(tea.KeyMsg)

	if !ok {
		return false
	}

	return key.Matches(keyMsg, bindings...)
}
