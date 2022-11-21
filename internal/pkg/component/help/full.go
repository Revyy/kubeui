package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Full returns a full help view.
func Full(width int, keys [][]key.Binding) string {
	h := New()
	h.Width = width
	return h.FullHelpView(keys)
}
