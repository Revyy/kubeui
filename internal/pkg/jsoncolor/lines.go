// Package jsoncolor provides a way to render a jsonString with syntax highlighting in the terminal.
package jsoncolor

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
)

// JSONLines builds colored json log lines.
// maxWidth defines the max width of a line in the result string before a linebreak is added.
func JSONLines(maxWidth int, jsonStr string) []string {
	maxWidthStyle := lipgloss.NewStyle().Width(maxWidth)

	formatter := NewFormatter()
	formatter.StringMaxLength = maxWidth * 10

	return slices.Filter(slices.Map(strings.Split(jsonStr, "\n"), func(str string) string {
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(str), &obj)
		if err != nil {
			return maxWidthStyle.Render(str)
		}

		s, err := formatter.Marshal(obj)
		if err != nil {
			return maxWidthStyle.Render(str)
		}

		return maxWidthStyle.Render(string(s))
	}), func(s string) bool {
		return len(s) > 0
	})
}
