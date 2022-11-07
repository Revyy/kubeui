package errorinfo

import (
	"kubeui/internal/pkg/kubeui"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
func New(message string) View {
	return View{
		keys:    newKeyMap(),
		message: message,
	}
}

// View displays an error and allows the user to quit the app.
type View struct {
	keys    *keyMap
	message string
}

// Update handles new messages from the runtime.
func (v View) Update(c kubeui.Context, msg kubeui.Msg) (kubeui.Context, kubeui.View, tea.Cmd) {

	if msg.MatchesKeyBindings(v.keys.Quit) {
		return c, v, kubeui.Exit()
	}

	return c, v, nil
}

// View renders the ui of the view.
func (v View) View(c kubeui.Context) string {
	builder := strings.Builder{}
	builder.WriteString(kubeui.ShortHelp(c.WindowWidth, []key.Binding{v.keys.Quit}))
	builder.WriteString("An error occured\n\n")
	builder.WriteString(kubeui.ErrorMessageStyle.Render(kubeui.LineBreak(v.message, c.WindowWidth)))

	return builder.String()
}

// Init initializes the view.
func (v View) Init(c kubeui.Context) tea.Cmd {
	return nil
}

// Destroy is called before a view is removed as the active view in the application.
func (v View) Destroy(c kubeui.Context) tea.Cmd {
	return nil
}
