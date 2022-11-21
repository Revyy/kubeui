package errorinfo

import (
	"kubeui/internal/pkg/component/help"
	"kubeui/internal/pkg/kubeui"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// keyMap defines the keys that are handled by this view.
type keyMap struct {
	kubeui.GlobalKeyMap
	Continue key.Binding
}

// newKeyMap defines the actual key bindings and creates a keyMap.
func newKeyMap() *keyMap {
	return &keyMap{
		GlobalKeyMap: kubeui.NewGlobalKeyMap(),
		Continue: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter,space", "Continue running the program"),
		),
	}
}

// New creates a new View.
func New(message string, windowWidth, windowHeight int) View {
	return View{
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
		keys:         newKeyMap(),
		message:      message,
	}
}

// View displays an error and allows the user to quit the app.
type View struct {
	windowWidth  int
	windowHeight int

	keys    *keyMap
	message string
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

	if msg.MatchesKeyBindings(v.keys.Quit) {
		return c, v, kubeui.Exit()
	}

	if msg.MatchesKeyBindings(v.keys.Continue) {
		return c, v, kubeui.PopView(false)
	}

	return c, v, nil
}

// View renders the ui of the view.
func (v View) View(c kubeui.Context) string {
	builder := strings.Builder{}
	builder.WriteString(help.Short(v.windowWidth, []key.Binding{v.keys.Quit, v.keys.Continue}))
	builder.WriteString("\n\n")
	builder.WriteString("An error occured\n\n")
	builder.WriteString(kubeui.ErrorMessageStyle.Render(kubeui.LineBreak(v.message, v.windowWidth)))

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
