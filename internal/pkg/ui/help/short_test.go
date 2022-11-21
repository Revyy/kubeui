package help_test

import (
	"kubeui/internal/pkg/ui/help"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/stretchr/testify/assert"
)

func TestShort(t *testing.T) {
	tests := []struct {
		name  string
		width int
		keys  []key.Binding
		want  string
	}{
		{"Width 0 should give empty string", 0, []key.Binding{key.NewBinding(key.WithKeys("esc"))}, ""},
		{"No keys should give empty string", 10, []key.Binding{}, ""},
		{"Should handle one key", 50, []key.Binding{key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Escape"))}, "esc Escape"},
		{"Should handle multiple keys", 50, []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Escape")),
			key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "Print")),
			key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Copy")),
		}, "esc Escape • ctrl+p Print • ctrl+c Copy"},
		{"Should truncate gracefully", 13, []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Escape")),
			key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "Print")),
			key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Copy")),
		}, "esc Escape …"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := help.Short(tt.width, tt.keys)
			assert.Equal(t, tt.want, got)
		})
	}
}
