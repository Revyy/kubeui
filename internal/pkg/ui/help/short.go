package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Short returns a short help view in a single line.
// If the line is wider than the specified width it will be gracefully truncated.
func Short(width int, keys []key.Binding) string {

	if width == 0 || len(keys) == 0 {
		return ""
	}

	h := New()
	h.Width = width
	return h.ShortHelpView(keys)
}
