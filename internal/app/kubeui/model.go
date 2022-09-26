package kubeui

import (
	"context"
	"kubeui/internal/app/kubeui/navigator"
	"kubeui/internal/app/kubeui/page"
	"kubeui/internal/app/kubeui/store"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

// model represents the base model
type Model struct {
	keys  *appKeyMap
	store *store.Store
	nav   *navigator.Navigator
}

func NewModel(kubernetesClientSet *kubernetes.Clientset, logger *zap.Logger) (tea.Model, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	namespaces, err := kubernetesClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	st := store.New(kubernetesClientSet).SetLogger(logger).SetNamespaces(namespaces.Items)

	nsSelectorConstructor := func(s *store.Store, paramaters map[string]string) page.Page {
		return page.NewNameSpaceSelector(s.Namespaces)
	}

	nav := navigator.New(nsSelectorConstructor(st, nil))
	nav.Add("namespace-selector", nsSelectorConstructor)

	return Model{
		keys:  newAppKeyMap(),
		store: st,
		nav:   nav,
	}, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.store.SetWindowSize(msg)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		}

	case error:
		return m, tea.Quit
	}

	cmd := m.nav.Current().Update(msg, m.store)

	return m, cmd

}

func newError(err error) tea.Cmd {
	return func() tea.Msg { return err }
}

func (m Model) View() string {
	view := m.nav.Current().View(m.store) + "\n"
	return view
}

func (m Model) Init() tea.Cmd {
	return m.nav.Current().Init(m.store)
}
