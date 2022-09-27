package confirm

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedButtonStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "200", Dark: "200"})
)

type keyMap struct {
	left  key.Binding
	right key.Binding
	enter key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "move one page to the left, if search mode is activated then move input cursor one position to the left"),
		),
		right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "move one page to the right, if search mode is activated then move input cursor one position to the right"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "press a button"),
		),
	}
}

// ButtonPress represents the action of pressing a button.
type ButtonPress struct {
	Button string
}

type Dialog struct {
	keys    *keyMap
	cursor  int
	buttons []string
	text    string
}

func New(buttons []string, text string) Dialog {

	return Dialog{
		keys:    newKeyMap(),
		buttons: buttons,
		cursor:  0,
		text:    text,
	}
}

func (d Dialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch {

		case key.Matches(msg, d.keys.left):
			if d.cursor > 0 {
				d.cursor--
			}

		case key.Matches(msg, d.keys.right):
			if d.cursor < len(d.buttons)-1 {
				d.cursor++
			}

		case key.Matches(msg, d.keys.enter):
			button := d.buttons[d.cursor]
			return d, func() tea.Msg {
				return ButtonPress{
					Button: button,
				}
			}

		}
	}
	return d, nil
}

func (d Dialog) View() string {
	// The header
	var dialogBuilder strings.Builder

	dialogBuilder.WriteString(d.text + "\n\n")

	for i, button := range d.buttons {
		if i == d.cursor {
			dialogBuilder.WriteString(selectedButtonStyle.Render(button))
		} else {
			dialogBuilder.WriteString(button)
		}

		if i < len(d.buttons)-1 {
			dialogBuilder.WriteString("   ")
		}
	}

	return dialogBuilder.String()
}

func (d Dialog) Init() tea.Cmd {
	return func() tea.Msg { return "" }
}
