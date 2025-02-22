package cxs

import (
	"fmt"
	"sort"
	"strings"

	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8s/k8scontext"
	"kubeui/internal/pkg/k8smsg"
	"kubeui/internal/pkg/ui/help"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// appKeyMap defines the keys that are handled at the top level in the application.
// These keys will be checked before passing along a msg to underlying components.
type appKeyMap struct {
	quit key.Binding
	help key.Binding
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
	}
}

// selectedContext is a type used to represent the selected context.
type selectedContext string

func (s selectedContext) String() string {
	return string(s)
}

// Model defines the base Model of the application.
type Model struct {
	// Client for manipulating kube-contexts.
	contextClient k8scontext.Client

	// application level keybindings
	keys *appKeyMap

	// searchtable used to select and delete contexts.
	table searchtable.Model

	// Yes/No dialog.
	// If non nil then the dialog is considered to be active.
	// A new dialog is created when needed.
	activeDialog *confirm.Model

	// Windows size
	windowSize tea.WindowSizeMsg

	showFullHelp bool
}

// NewModel creates a new cxs model.
func NewModel(contextClient k8scontext.Client) *Model {
	contexts := contextClient.Contexts()
	sort.Strings(contexts)

	table := searchtable.New(contexts, 10, contextClient.CurrentContext(), true, searchtable.Options{SingularItemName: "context"})

	return &Model{
		keys:          newAppKeyMap(),
		contextClient: contextClient,
		table:         table,
	}
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keys.help, m.keys.quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.help, m.keys.quit},
		m.table.KeyList(),
	}
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can it can gracefully truncate
		// its view as needed.
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.help):
			m.showFullHelp = !m.showFullHelp
			return m, nil
		}

	case error:
		return m, tea.Quit

	case searchtable.Selection:

		return m, func() tea.Msg {
			err := m.contextClient.SwitchContext(msg.Value, "")
			if err != nil {
				return err
			}

			return selectedContext(msg.Value)
		}

	case selectedContext:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(searchtable.UpdateHighlighted{Item: msg.String()})
		return m, cmd

	case searchtable.Deletion:
		dialog := confirm.New([]confirm.Button{{Desc: "Yes", Id: msg.Value}, {Desc: "No", Id: msg.Value}}, fmt.Sprintf("Are you sure you want to delete %s", msg.Value))
		m.activeDialog = &dialog
		return m, nil
	case confirm.ButtonPress:

		// If the user pressed No then we close the dialog and reset contextToDelete.
		if msg.Pressed.Desc != "Yes" {
			m.activeDialog = nil
			return m, nil
		}

		return m, func() tea.Msg {
			err := deleteContext(msg.Pressed.Id, m.contextClient)
			if err != nil {
				return err
			}

			return k8smsg.NewContextDeletedMsg(msg.Pressed.Id)
		}

	case k8smsg.ContextDeletedMsg:
		items := []string{}
		for _, k := range m.contextClient.Contexts() {
			if k != msg.Name {
				items = append(items, k)
			}
		}
		sort.Strings(items)

		m.activeDialog = nil
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(searchtable.UpdateItems{Items: items})
		return m, cmd
	}

	if m.activeDialog != nil {
		dialog, cmd := m.activeDialog.Update(msg)
		m.activeDialog = &dialog
		return m, cmd
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

// deleteContext deletes a kubernetes context and the corresponding cluster entry and user entry.
func deleteContext(kubeCtx string, contextClient k8scontext.Client) error {
	err := contextClient.DeleteContext(kubeCtx)
	if err != nil {
		return err
	}

	err = contextClient.DeleteClusterEntry(kubeCtx)
	if err != nil {
		return err
	}

	err = contextClient.DeleteUser(kubeCtx)

	return err
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (m Model) View() string {
	builder := strings.Builder{}

	helpView := help.Short(m.windowSize.Width, m.ShortHelp())

	if m.showFullHelp {
		helpView = help.Full(m.windowSize.Width, m.FullHelp())
	}

	builder.WriteString(helpView)
	builder.WriteString("\n\n")

	if m.activeDialog != nil {
		builder.WriteString(m.activeDialog.View())
		return builder.String()
	}

	builder.WriteString(m.table.View())

	return builder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (m Model) Init() tea.Cmd {
	return nil
}
