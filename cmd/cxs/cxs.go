package main

import (
	"flag"
	"fmt"
	"kubeui/internal/app/kubeui"
	"kubeui/internal/pkg/component/confirm"
	"kubeui/internal/pkg/component/searchtable"
	"kubeui/internal/pkg/k8s"
	"log"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// appKeyMap defines the keys that are handled at the top level in the application.
// These keys will be checked before passing along a msg to underlying components.
type appKeyMap struct {
	quit key.Binding
}

// newAppKeyMap defines the actual key bindings and creates an appKeyMap.
func newAppKeyMap() *appKeyMap {
	return &appKeyMap{
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c, ctrl+q", "quit the app"),
		),
	}
}

// model defines the base model of the application.
type model struct {
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
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		}

	case error:
		return m, tea.Quit

	case searchtable.Selection:
		err := k8s.SwitchContext(msg.Value, m.configAccess, m.config)
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

func (m model) View() string {

	if m.activeDialog != nil {
		return m.activeDialog.View()
	}

	return m.table.View()
}

func (m model) Init() tea.Cmd {
	return nil
}

func main() {

	// If a specific kubeconfig file is specified then we load that, otherwise the defaults will be loaded.
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	clientConfig := k8s.NewClientConfig("", *kubeconfig)

	rawConfig, err := clientConfig.RawConfig()

	if err != nil {
		log.Fatalf("failed to load config")
	}

	configAccess := clientConfig.ConfigAccess()

	items := []string{}

	for k := range rawConfig.Contexts {
		items = append(items, k)
	}

	sort.Strings(items)

	table := searchtable.New(items, 10, rawConfig.CurrentContext, true)

	m := model{
		keys:         newAppKeyMap(),
		config:       rawConfig,
		configAccess: configAccess,
		table:        table,
	}

	program := kubeui.NewProgram(m)
	kubeui.StartProgram(program)

}
