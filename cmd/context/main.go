package main

import (
	"flag"
	"kubeui/internal/app/kubeui"
	"kubeui/internal/pkg/k8s"
	"log"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
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
	keys   *appKeyMap
	config api.Config
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
	}

	return m, nil

}

func (m model) View() string {
	return "hello"
}

func (m model) Init() tea.Cmd {
	return nil
}

func main() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	rawConfig, err := k8s.NewRawConfig("", *kubeconfig)

	if err != nil {
		log.Fatalf("failed to load config")
	}

	m := model{
		keys:   newAppKeyMap(),
		config: rawConfig,
	}

	program := kubeui.NewProgram(m)
	kubeui.StartProgram(program)

}
