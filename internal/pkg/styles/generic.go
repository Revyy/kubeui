// Package styles contains lipgloss styles that can be reused.
package styles

import "github.com/charmbracelet/lipgloss"

// ErrorMessage is used to style error messages.
var ErrorMessage = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9"))
