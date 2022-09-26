package page

import (
	"kubeui/internal/app/kubeui/store"

	tea "github.com/charmbracelet/bubbletea"
)

type Page interface {
	Update(msg tea.Msg, state *store.Store) tea.Cmd
	Init(*store.Store) tea.Cmd
	View(*store.Store) string
}
