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

// KeyMap defines the key bindings for the dialog.
type KeyMap struct {
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
}

// newKeyMap creates a new KeyMap.
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

// Button represents a button.
type Button struct {
	Id   string
	Desc string
}

// ButtonPress represents the action of pressing a button.
type ButtonPress struct {
	Pressed Button
}

// Model defines a component use to confirm a choice.
type Model struct {
	keys    *KeyMap
	cursor  int
	buttons []Button
	text    string
}

// Returns a list of keybindings to be used in help text.
func (d Model) KeyList() []key.Binding {
	keyList := []key.Binding{
		d.keys.Left,
		d.keys.Right,
		d.keys.Enter,
	}

	return keyList
}

// New creates a new Model.
func New(buttons []Button, text string) Model {

	return Model{
		keys:    newKeyMap(),
		buttons: buttons,
		cursor:  0,
		text:    text,
	}
}

// Update updates the model and optionally returns a command.
// It is part of the bubbletea model interface.
func (d Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, d.keys.Left):
			if d.cursor > 0 {
				d.cursor--
			}

		case key.Matches(msg, d.keys.Right):
			if d.cursor < len(d.buttons)-1 {
				d.cursor++
			}

		case key.Matches(msg, d.keys.Enter):
			button := d.buttons[d.cursor]
			return d, func() tea.Msg {
				return ButtonPress{
					Pressed: button,
				}
			}

		}
	}
	return d, nil
}

// View returns the view for the model.
// It is part of the bubbletea model interface.
func (d Model) View() string {
	// The header
	var dialogBuilder strings.Builder

	dialogBuilder.WriteString(d.text + "\n\n")

	for i, button := range d.buttons {
		if i == d.cursor {
			dialogBuilder.WriteString(selectedButtonStyle.Render(button.Desc))
		} else {
			dialogBuilder.WriteString(button.Desc)
		}

		if i < len(d.buttons)-1 {
			dialogBuilder.WriteString("   ")
		}
	}

	return dialogBuilder.String()
}

// Init returns an initial command.
// It is part of the bubbletea model interface.
func (d Model) Init() tea.Cmd {
	return func() tea.Msg { return "" }
}
