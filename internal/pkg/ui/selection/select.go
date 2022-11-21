package selection

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
)

// selectedStyle is used for selected items.
var selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})

// UnselectedStyle is used for unselected items.
var unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})

// Tabs renders a tab select.
func Tabs(cursor, maxWidth int, headers []string) string {
	var tabsBuilder strings.Builder

	// Iterate over the items in the current page and print them out.
	for i, header := range headers {

		// Is the cursor pointing at this choice?
		if cursor == i {
			tabsBuilder.WriteString(lipgloss.NewStyle().Underline(true).Render(header) + " ")
			continue
		}

		tabsBuilder.WriteString(header + " ")
	}

	return lipgloss.NewStyle().Width(maxWidth).Render(tabsBuilder.String())
}

// HorizontalList renders a horizontal list with an item highlighted as selected.
func HorizontalList(items []string, selectedItem string, maxWidth int) string {

	builder := strings.Builder{}

	for i, item := range items {
		if item == selectedItem {
			builder.WriteString(selectedStyle.Render(fmt.Sprintf("[%d] %s", i+1, item)) + " ")
			continue
		}
		builder.WriteString(unselectedStyle.Render(fmt.Sprintf("[%d] %s", i+1, item)) + " ")
	}

	return wrap.String(builder.String(), maxWidth)
}
