package kubeui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var HelpBolderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63"))

// LineBreak splits a string into multiple lines to make each line a max of maxWidth in length.
func LineBreak(str string, maxWidth int) string {

	if len(str) <= maxWidth {
		return str
	}

	if len(strings.Split(str, " ")) == 1 && len(str) > maxWidth {
		return str[:maxWidth]
	}

	charCount := 0
	newStr := []rune{}

	var i int
	for i < len(str) {
		char := str[i]
		// If we have not yet reached the charCount.
		if charCount < maxWidth-1 {
			charCount++
			newStr = append(newStr, rune(char))
			i++
			continue
		}

		// If we have reached the charCount and we have a space, then we replace the space with a newline.
		if str[i] == ' ' {
			newStr = append(newStr, '\n')
			charCount = 0
			i++
			continue
		}

		newStr = append(newStr, '\n')
		newStr = append(newStr, rune(char))
		charCount = 0
		i++
	}
	return string(newStr)

}

// Truncate truncates a string and adds ... to indicate it was truncated.
// It keeps words intact as falls back to the last word with room for ...
func Truncate(text string, maxLen int) string {

	maxIdx := maxLen - 3

	if maxIdx <= 0 {
		return ""
	}

	lastSpace := maxIdx
	len := 0
	for i, r := range text {
		if unicode.IsSpace(r) {
			lastSpace = i
		}
		len++
		if len > maxIdx {
			return text[:lastSpace] + "..."
		}
	}

	return text
}
