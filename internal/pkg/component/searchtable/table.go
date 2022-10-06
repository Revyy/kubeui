package searchtable

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedPageStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"})
	unSelectedPageStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})
	highlightedStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "200", Dark: "200"})
)

type KeyMap struct {
	Search     key.Binding
	ExitSearch key.Binding
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Enter      key.Binding
	Delete     key.Binding
}

// Selection represents the act of selecting an item.
type Selection struct {
	Value string
}

// Deletion represents the act of deleting an item.
type Deletion struct {
	Value string
}

// UpdateItems resets the base items list for the table to the items passed in.
type UpdateItems struct {
	Items []string
}

// UpdateHighlighted sets a new previous choice values.
type UpdateHighlighted struct {
	Item string
}

func newKeyMap(itemName string) *KeyMap {

	itemName = strings.ToLower(itemName)

	selectPhrase := "Select an item"

	if itemName != "" {
		selectPhrase = fmt.Sprintf("Select a %s", itemName)
	}

	deletePhrase := "Delete an item"

	if itemName != "" {
		deletePhrase = fmt.Sprintf("Delete a %s", itemName)
	}

	return &KeyMap{
		Search: key.NewBinding(
			key.WithKeys("ctrl+s", "cmd+f", "ctrl+f"),
			key.WithHelp("ctrl+s,cmd+f,ctrl+f", "Enter search mode"),
		),
		ExitSearch: key.NewBinding(
			key.WithKeys("ctrl+s", "cmd+f", "enter", "esc", "down"),
			key.WithHelp("ctrl+s,cmd+f,enter,esc,down", "Exit search mode"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "Move cursor up one position"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "Move cursor down one position"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "Move one page or the cursor to the left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "Move one page or the cursor to the right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", selectPhrase),
		),
		Delete: key.NewBinding(
			key.WithKeys("delete"),
			key.WithHelp("delete", deletePhrase),
		),
	}
}

// Options specifies additional options to be considered when creating a searchtable.
type Options struct {
	// Used to modify help texts for keys.
	SingularItemName string
}

type SearchTable struct {
	keys   *KeyMap
	cursor int
	items  []string

	highlighted string

	allowDelete       bool
	currentItemsSlice []string
	currentPage       int
	pageSize          int
	numPages          int
	numItems          int

	numFilteredItems int
	searchField      textinput.Model
	searchMode       bool
}

// Returns a list of keybindings to be used in help text.
func (st SearchTable) KeyList() []key.Binding {
	keyList := []key.Binding{
		st.keys.Search,
		st.keys.ExitSearch,
		st.keys.Up,
		st.keys.Down,
		st.keys.Left,
		st.keys.Right,
		st.keys.Enter,
	}

	if st.allowDelete {
		keyList = append(keyList, st.keys.Delete)
	}

	return keyList
}

func calcSlice(length, currentPage, pageSize int) (int, int) {
	if pageSize == 0 {
		return 0, 0
	}
	// Should not happen
	if currentPage*pageSize > length {
		return 0, 0
	}

	if currentPage*pageSize+pageSize > length {
		return currentPage * pageSize, length
	}

	return currentPage * pageSize, currentPage*pageSize + pageSize
}

func New(items []string, pageSize int, previousChoice string, allowDelete bool, options Options) SearchTable {
	searchField := textinput.New()
	searchField.Placeholder = ""
	searchField.Focus()
	searchField.CharLimit = 156
	searchField.Width = 20

	numPages := int(math.Ceil(float64(len(items)) / float64(pageSize)))
	numItems := len(items)

	sliceStart, sliceEnd := calcSlice(numItems, 0, pageSize)

	return SearchTable{
		keys:              newKeyMap(options.SingularItemName),
		items:             items,
		currentItemsSlice: items[sliceStart:sliceEnd],
		allowDelete:       allowDelete,
		highlighted:       previousChoice,
		pageSize:          pageSize,
		numPages:          numPages,
		searchField:       searchField,
		numItems:          numItems,
		numFilteredItems:  numItems,
	}
}

func (st SearchTable) Update(msg tea.Msg) (SearchTable, tea.Cmd) {

	var cmd tea.Cmd

	if st.searchMode {
		st, cmd = updateInSearchMode(st, msg)
	} else {
		st, cmd = updateInselectMode(st, msg)
		if cmd != nil {
			return st, cmd
		}
	}

	switch m := msg.(type) {
	case UpdateItems:
		st.items = m.Items
	case UpdateHighlighted:
		st.highlighted = m.Item
	}

	// Filter items based on the search value.
	filteredItems := []string{}

	for _, item := range st.items {
		if strings.Contains(item, st.searchField.Value()) {
			filteredItems = append(filteredItems, item)
		}
	}

	// If we have a search result that is different than the last result we reset the page.
	if numFilteredItems := len(filteredItems); numFilteredItems != st.numFilteredItems {
		st.numFilteredItems = numFilteredItems
		st.currentPage = 0
		st.numPages = int(math.Ceil(float64(st.numFilteredItems) / float64(st.pageSize)))
	}

	// Calculate which items should be displayed based on the current page and the pageSize.
	sliceStart, sliceEnd := calcSlice(st.numFilteredItems, st.currentPage, st.pageSize)
	st.currentItemsSlice = filteredItems[sliceStart:sliceEnd]

	// If the selection on the previous page was at a higher index than the current pages total items
	// then we reset it to avoid having a missing cursor.
	if st.cursor > len(st.currentItemsSlice)-1 {
		st.cursor = 0
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return st, cmd

}

func updateInselectMode(st SearchTable, msg tea.Msg) (SearchTable, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch {

		case key.Matches(msg, st.keys.Search):
			st.searchMode = true
			return st, nil
		// The "up" and "k" keys move the cursor up
		case key.Matches(msg, st.keys.Up):
			if st.cursor <= 0 {
				st.searchMode = true
				return st, nil
			}
			st.cursor--

		// The "down" and "j" keys move the cursor down
		case key.Matches(msg, st.keys.Down):
			if st.cursor < len(st.currentItemsSlice)-1 {
				st.cursor++
			}

		case key.Matches(msg, st.keys.Left):
			if st.currentPage > 0 {
				st.currentPage--
			}

		case key.Matches(msg, st.keys.Right):
			if st.currentPage < st.numPages-1 {
				st.currentPage++
			}

		case key.Matches(msg, st.keys.Enter):
			item := st.currentItemsSlice[st.cursor]
			return st, func() tea.Msg {
				return Selection{Value: item}
			}
		case key.Matches(msg, st.keys.Delete) && st.allowDelete:
			item := st.currentItemsSlice[st.cursor]
			return st, func() tea.Msg {
				return Deletion{Value: item}
			}
		}
	}

	return st, nil
}

func updateInSearchMode(st SearchTable, msg tea.Msg) (SearchTable, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, st.keys.ExitSearch):
			st.searchMode = false
			return st, nil
		}
	}
	st.searchField, _ = st.searchField.Update(msg)
	return st, nil
}

func (n SearchTable) View() string {

	var mainBuilder strings.Builder

	searchStyle := selectedPageStyle
	if !n.searchMode {
		searchStyle = unSelectedPageStyle
	}

	mainBuilder.WriteString(searchStyle.Render(n.searchField.View()) + "\n\n\n")

	// The header
	var selectBuilder strings.Builder

	// Iterate over the namespaces in the current page and print them out.
	for i, item := range n.currentItemsSlice {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if n.cursor == i {
			cursor = ">" // cursor!
		}
		// Render the row
		if item == n.highlighted {
			//selectBuilder.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
			selectBuilder.WriteString(fmt.Sprintf("%s %s\n", cursor, highlightedStyle.Render(item)))
		} else {
			selectBuilder.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
		}

	}

	// Start building the pageinator view.
	paginatorView := "\n\n"
	// If we are not at the first page then display a left arrow indicating that we can go left.
	if n.currentPage > 0 {
		paginatorView += "< "
	} else { // Else just print space to fill the void of the arrow.
		paginatorView += "  "
	}

	// Print out the pages in order as [1 2 3 4] etc.
	paginatorView += "[ "
	for i := 0; i < n.numPages; i++ {
		if i == n.currentPage {
			paginatorView += selectedPageStyle.Render(fmt.Sprintf("%d", i+1))
		} else {
			paginatorView += unSelectedPageStyle.Render(fmt.Sprintf("%d", i+1))
		}
		// Add space between numbers inside the brackets.
		if i < n.numPages-1 {
			paginatorView += "  "
		}
	}
	paginatorView += " ]"

	// If we are not at the last page then display a right arrow indicating that we can go right.
	if n.currentPage < n.numPages-1 {
		paginatorView += " >"
	}

	selectBuilder.WriteString(paginatorView)

	selectStyle := selectedPageStyle
	if n.searchMode {
		selectStyle = unSelectedPageStyle
	}

	mainBuilder.WriteString(selectStyle.Render(selectBuilder.String()))

	return mainBuilder.String()
}

func (n SearchTable) Init() tea.Cmd {
	return func() tea.Msg { return "" }
}
