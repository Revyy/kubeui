package cxs

import (
	"fmt"
	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8s"
	"sort"
	"strings"

	"kubeui/internal/pkg/kubeui/help"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
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

func NewModel(rawConfig api.Config, configAccess clientcmd.ConfigAccess) *Model {

	items := []string{}

	for k := range rawConfig.Contexts {
		items = append(items, k)
	}

	sort.Strings(items)

	table := searchtable.New(items, 10, rawConfig.CurrentContext, true)

	return &Model{
		keys:         newAppKeyMap(),
		config:       rawConfig,
		configAccess: configAccess,
		table:        table,
		help:         help.New(),
	}
}

// Model defines the base Model of the application.
type Model struct {
	// application level keybindings
	keys *appKeyMap

	// kubernetes config object.
	config api.Config

	// object defining how the kubernetes config was located and put together.
	// needed in order to modify the config files on disc.
	configAccess clientcmd.ConfigAccess

	// searchtable used to select and delete contexts.
	table searchtable.SearchTable

	// Yes/No dialog.
	// If non nil then the dialog is considered to be active.
	// A new dialog is created when needed.
	activeDialog *confirm.Dialog

	// Used to store the id of the context to delete.
	// Needed as we use a confirm dialog to get approval before deleting which
	// means that we need to do it across multiple calls to the Update function.
	contextToDelete string

	// Windows size
	windowSize tea.WindowSizeMsg
	// Help
	help help.Model
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
		{
			m.table.Keys.Up,
			m.table.Keys.Left,
			m.table.Keys.Right,
			m.table.Keys.Down,
			m.table.Keys.Enter,
			m.table.Keys.Delete,
			m.table.Keys.Search,
			m.table.Keys.ExitSearch,
		},
	}
}

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
		err := k8s.SwitchContext(msg.Value, "", m.configAccess, m.config)
		if err != nil {
			return m, func() tea.Msg { return err }
		}
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(searchtable.UpdateHighlighted{Item: msg.Value})
		return m, cmd
	case searchtable.Deletion:
		m.contextToDelete = msg.Value
		dialog := confirm.New([]string{"Yes", "No"}, fmt.Sprintf("Are you sure you want to delete %s", msg.Value))
		m.activeDialog = &dialog
		return m, nil
	case confirm.ButtonPress:

		// If the user pressed No then we close the dialog and reset contextToDelete.
		if msg.Button != "Yes" {
			m.contextToDelete = ""
			m.activeDialog = nil
			return m, nil
		}

		err := k8s.DeleteContext(m.contextToDelete, m.configAccess, m.config)

		if err != nil {
			return m, tea.Quit
		}

		err = k8s.DeleteClusterEntry(m.contextToDelete, m.configAccess, m.config)

		if err != nil {
			return m, tea.Quit
		}

		err = k8s.DeleteUser(m.contextToDelete, m.configAccess, m.config)

		if err != nil {
			return m, tea.Quit
		}

		items := []string{}

		for k := range m.config.Contexts {
			if k != m.contextToDelete {
				items = append(items, k)
			}
		}
		sort.Strings(items)

		m.contextToDelete = ""
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

func (m Model) Init() tea.Cmd {
	return nil
}
