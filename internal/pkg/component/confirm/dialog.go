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

type KeyMap struct {
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
}

func newKeyMap() *KeyMap {
	return &KeyMap{
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "Select the button to the left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "Select the button to the right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Make a choice"),
		),
	}
}

// ButtonPress represents the action of pressing a button.
type ButtonPress struct {
	Button string
}

type Dialog struct {
	Keys    *KeyMap
	cursor  int
	buttons []string
	text    string
}

func New(buttons []string, text string) Dialog {

	return Dialog{
		Keys:    newKeyMap(),
		buttons: buttons,
		cursor:  0,
		text:    text,
	}
}

func (d Dialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, d.Keys.Left):
			if d.cursor > 0 {
				d.cursor--
			}

		case key.Matches(msg, d.Keys.Right):
			if d.cursor < len(d.buttons)-1 {
				d.cursor++
			}

		case key.Matches(msg, d.Keys.Enter):
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
