package cxs

import (
	"fmt"
	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/kubeui"
	"sort"
	"strings"

	"kubeui/internal/pkg/component/help"

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

// Model defines the base Model of the application.
type Model struct {

	//
	contextClient kubeui.ContextClient

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
	// Help
	help help.Model
}

// NewModel creates a new cxs model.
func NewModel(contextClient kubeui.ContextClient) *Model {

	contexts := contextClient.Contexts()
	sort.Strings(contexts)

	table := searchtable.New(contexts, 10, contextClient.CurrentContext(), true, searchtable.Options{SingularItemName: "context"})

	return &Model{
		keys:          newAppKeyMap(),
		contextClient: contextClient,
		table:         table,
		help:          help.New(),
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
		m.help.Width = msg.Width
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

	case error:
		return m, tea.Quit

	case searchtable.Selection:
		err := m.contextClient.SwitchContext(msg.Value, "")
		if err != nil {
			return m, func() tea.Msg { return err }
		}
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(searchtable.UpdateHighlighted{Item: msg.Value})
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

		err := deleteContext(msg.Pressed.Id, m.contextClient)

		if err != nil {
			return m, tea.Quit
		}

		items := []string{}
		for _, k := range m.contextClient.Contexts() {
			if k != msg.Pressed.Id {
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
func deleteContext(kubeCtx string, contextClient kubeui.ContextClient) error {
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

	helpView := m.help.View(m)
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
