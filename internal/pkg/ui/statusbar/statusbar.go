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
	statusBar := statusBarStyle.Width(width)

	builder := &strings.Builder{}

	for _, v := range values {
		builder.WriteString(fmt.Sprintf("%s%s", v, separator))
	}

	return statusBar.Render(lipgloss.NewStyle().Width(width).Render(strings.TrimRight(builder.String(), separator)))
}
