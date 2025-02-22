// Package statusbar provides a stateless component representing a status bar.
// The status bar should be used to provide metadata that is useful to the user for a specific view.
package statusbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarStyle is used to create a status bar displaying key information about the running app.
var statusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("201")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	BorderBottom(true)

// New returns a string representing a status bar.
// Values are printed with separator in between.
func New(width int, separator string, values ...string) string {
	if width <= 0 {
		return ""
	}

	builder := &strings.Builder{}

	for _, v := range values {
		builder.WriteString(fmt.Sprintf("%s%s", v, separator))
	}

	return lipgloss.NewStyle().Width(width).Render(statusBarStyle.Render(strings.TrimRight(builder.String(), separator)))
}
