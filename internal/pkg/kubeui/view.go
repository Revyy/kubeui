package kubeui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

// Context contains the context of the kubeui application.
type Context struct {
	WindowWidth  int
	WindowHeight int
	// object defining how the kubernetes config was located and put together.
	// needed in order to modify the config files on disc.
	ConfigAccess clientcmd.ConfigAccess

	// ClientSet used to issue commands to kubernetes.
	Kubectl *kubernetes.Clientset

	// kubernetes config object.
	ApiConfig api.Config

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
