// Package selection provides functions to render ui components used for selection.
// The logic of actually selecting something and keeping track of which item is selected is not provided by this package.
// The selection package simply provides the tools to render such ui components.
package selection

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// selectedStyle is used for selected items.
var selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})

// unselectedStyle is used for unselected items.
var unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "252", Dark: "235"})

// underlineStyle is used to underline text.
var underlineStyle = lipgloss.NewStyle().Underline(true)

// Tabs renders a tab select.
// cursor indicates the selected index in headers.
// The selected header is rendered using a highlighted color style and will be underlined.
func Tabs(cursor, maxWidth int, headers []string) string {
	if maxWidth == 0 {
		return ""
	}

	var tabsBuilder strings.Builder

	// Iterate over the items in the current page and print them out.
	for i, header := range headers {

		// Is the cursor pointing at this choice?
		if cursor == i {
			tabsBuilder.WriteString(underlineStyle.Render(header) + " ")
			continue
		}

		tabsBuilder.WriteString(header + " ")
	}

	return lipgloss.NewStyle().Width(maxWidth).Render(strings.Trim(tabsBuilder.String(), " "))
}

// HorizontalList renders a horizontal list with an item highlighted as selected.
// Example: [1] Item 1 [2] Item 2 [3] Item 3.
func HorizontalList(items []string, selectedItem string, maxWidth int) string {
	if maxWidth == 0 || len(items) == 0 {
		return ""
	}

	builder := strings.Builder{}

	for i, item := range items {
		if item == selectedItem {
			builder.WriteString(selectedStyle.Render(fmt.Sprintf("[%d] %s", i+1, item)) + " ")
			continue
		}
		builder.WriteString(unselectedStyle.Render(fmt.Sprintf("[%d] %s", i+1, item)) + " ")
	}

	return lipgloss.NewStyle().Width(maxWidth).Render(strings.Trim(builder.String(), " "))
}
