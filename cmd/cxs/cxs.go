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

type appKeyMap struct {
	quit key.Binding
}

func newAppKeyMap() *appKeyMap {
	return &appKeyMap{
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c, ctrl+q", "quit the app"),
		),
	}
}

type model struct {
	keys            *appKeyMap
	config          api.Config
	configAccess    clientcmd.ConfigAccess
	table           searchtable.SearchTable
	activeDialog    *confirm.Dialog
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
		return m, tea.Quit
	case searchtable.Deletion:
		m.contextToDelete = msg.Value
		dialog := confirm.New([]string{"Yes", "No"}, fmt.Sprintf("Are you sure you want to delete %s", msg.Value))
		m.activeDialog = &dialog
		return m, nil
	case confirm.ButtonPress:

		if msg.Button != "Yes" {
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

		return m, func() tea.Msg { return searchtable.UpdateItems{Items: items} }
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
