package namespaceselection

import (
	"fmt"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/k8scommand"
	"kubeui/internal/pkg/kubeui"
	"strings"

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

// New creates a new View.
func New() View {
	return View{
		keys: newKeyMap(),
	}
}

// View allows the user to select a namespace.
type View struct {
	keys *keyMap

	// Namespaces in current cluster.
	namespaces []string

	// SearchTable used to select a namespace.
	namespaceTable searchtable.Model
}

// Update handles new messages from the runtime.
func (v View) Update(c kubeui.Context, msg kubeui.Msg) (kubeui.Context, kubeui.View, tea.Cmd) {

	if msg.MatchesKeyBindings(v.keys.Quit) {
		return c, v, kubeui.Exit()
	}

	// Results
	switch t := msg.TeaMsg.(type) {
	case k8scommand.ListNamespacesMsg:
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
		return c, v, nil

	case searchtable.Selection:

		err := k8s.SwitchContext(c.ApiConfig.CurrentContext, t.Value, c.ConfigAccess, c.ApiConfig)
		if err != nil {
			return c, v, kubeui.Error(err)
		}

		c.Namespace = t.Value
		return c, v, kubeui.PushView("pod_selection")

	}

	var cmd tea.Cmd
	v.namespaceTable, cmd = v.namespaceTable.Update(msg.TeaMsg)
	return c, v, cmd
}

// View renders the ui of the view.
func (v View) View(c kubeui.Context) string {
	builder := strings.Builder{}

	builder.WriteString(kubeui.ShortHelp(c.WindowWidth, []key.Binding{v.keys.Quit}))
	builder.WriteString("\n\n")

	statusBar := kubeui.StatusBar(c.WindowWidth-1, " ", fmt.Sprintf("Context: %s", c.ApiConfig.CurrentContext))
	builder.WriteString(statusBar + "\n")

	builder.WriteString(v.namespaceTable.View())

	return builder.String()
}

// Init initializes the view.
func (v View) Init(c kubeui.Context) tea.Cmd {
	return k8scommand.ListNamespaces(c.Kubectl)
}

// Destroy is called before a view is removed as the active view in the application.
func (v View) Destroy(c kubeui.Context) tea.Cmd {
	return nil
}
