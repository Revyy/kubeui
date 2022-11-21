package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Short returns a short help view.
func Short(width int, keys []key.Binding) string {
	h := New()
	h.Width = width
	return h.ShortHelpView(keys)
}
