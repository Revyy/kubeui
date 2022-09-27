package page

import (
	"kubeui/internal/app/kubeui/store"
	"kubeui/internal/pkg/component/searchtable"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
)

type NameSpaceSelector struct {
	namespaces  []v1.Namespace
	searchTable searchtable.SearchTable
}

func NewNameSpaceSelector(namespaces []v1.Namespace) *NameSpaceSelector {

	items := []string{}

	for _, ns := range namespaces {
		items = append(items, ns.Name)
	}

	searchTable := searchtable.New(items, 5, "", false)

	return &NameSpaceSelector{
		namespaces:  namespaces,
		searchTable: searchTable,
	}
}

func (n *NameSpaceSelector) Update(msg tea.Msg, store *store.Store) tea.Cmd {
	var cmd tea.Cmd
	n.searchTable, cmd = n.searchTable.Update(msg)

	return cmd
}

func (n *NameSpaceSelector) View(store *store.Store) string {
	return n.searchTable.View()
}

func (n *NameSpaceSelector) Init(store *store.Store) tea.Cmd {
	return func() tea.Msg { return "" }
}
