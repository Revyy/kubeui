package help_test

import (
	"kubeui/internal/pkg/ui/help"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/stretchr/testify/assert"
)

func TestFull(t *testing.T) {

	tests := []struct {
		name  string
		width int
		keys  [][]key.Binding
		want  string
	}{
		{"Width 0 should give empty string", 0, [][]key.Binding{{key.NewBinding(key.WithKeys("esc"))}}, ""},
		{"No keys should give empty string", 10, [][]key.Binding{}, ""},
		{"Should handle rendering of columns", 100, [][]key.Binding{
			{key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Escape"))},
			{key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Copy"))},
			{key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "Print"))},
		}, "esc Escape    ctrl+c Copy    ctrl+p Print    "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := help.Full(tt.width, tt.keys)
			assert.Equal(t, tt.want, got)
		})
	}
}
