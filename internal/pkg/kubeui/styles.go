package kubeui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ErrorMessageStyle is used to style error messages.
var ErrorMessageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9"))

// StatusBarStyle is used to create a status bar displaying key information about the running app.
var StatusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("201")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	BorderBottom(true)

// StatusBar returns a string representing a status bar.
// Values are printed with separator in between.
func StatusBar(width int, separator string, values ...string) string {
	statusBar := StatusBarStyle.Width(width)

	builder := &strings.Builder{}

	for _, v := range values {
		builder.WriteString(fmt.Sprintf("%s%s", v, separator))
	}

	return statusBar.Render(lipgloss.NewStyle().Width(width).Render(strings.TrimRight(builder.String(), separator)))
}

// SelectedStyle is used for selected items.
var SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})

// UnselectedStyle is used for unselected items.
var UnselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})
