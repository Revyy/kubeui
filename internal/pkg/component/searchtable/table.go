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
)

type keyMap struct {
	search     key.Binding
	up         key.Binding
	down       key.Binding
	left       key.Binding
	right      key.Binding
	exitSearch key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		search: key.NewBinding(
			key.WithKeys("ctrl+s", "cmd+f", "ctrl+f"),
			key.WithHelp("ctrl+s, cmd+f, ctrl+f", "enter search mode"),
		),
		exitSearch: key.NewBinding(
			key.WithKeys("ctrl+s", "cmd+f", "enter", "esc", "down"),
			key.WithHelp("ctrl+s, cmd+f, enter, esc, down", "exit search mode"),
		),
		up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "move cursor up one position, if at top position then search mode will be activated"),
		),
		down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "move cursor down one position, if search mode is active then it will be deactivated"),
		),
		left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "move one page to the left, if search mode is activated then move input cursor one position to the left"),
		),
		right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "move one page to the right, if search mode is activated then move input cursor one position to the right"),
		),
	}
}

type SearchTable struct {
	keys   *keyMap
	cursor int
	items  []string

	currentItemsSlice []string
	currentPage       int
	pageSize          int
	numPages          int
	numItems          int

	numFilteredItems int
	searchField      textinput.Model
	searchMode       bool
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

func New(items []string, pageSize int) SearchTable {
	searchField := textinput.New()
	searchField.Placeholder = ""
	searchField.Focus()
	searchField.CharLimit = 156
	searchField.Width = 20

	numPages := int(math.Ceil(float64(len(items)) / float64(pageSize)))
	numItems := len(items)

	sliceStart, sliceEnd := calcSlice(numItems, 0, pageSize)

	return SearchTable{
		keys:              newKeyMap(),
		items:             items,
		currentItemsSlice: items[sliceStart:sliceEnd],
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

		case key.Matches(msg, st.keys.search):
			st.searchMode = true
			return st, nil
		// The "up" and "k" keys move the cursor up
		case key.Matches(msg, st.keys.up):
			if st.cursor <= 0 {
				st.searchMode = true
				return st, nil
			}
			st.cursor--

		// The "down" and "j" keys move the cursor down
		case key.Matches(msg, st.keys.down):
			if st.cursor < len(st.currentItemsSlice)-1 {
				st.cursor++
			}

		case key.Matches(msg, st.keys.left):
			if st.currentPage > 0 {
				st.currentPage--
			}

		case key.Matches(msg, st.keys.right):
			if st.currentPage < st.numPages-1 {
				st.currentPage++
			}
		}
	}

	return st, nil
}

func updateInSearchMode(st SearchTable, msg tea.Msg) (SearchTable, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, st.keys.exitSearch):
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
	selectBuilder.WriteString("Select a namespace\n\n")

	// Iterate over the namespaces in the current page and print them out.
	for i, item := range n.currentItemsSlice {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if n.cursor == i {
			cursor = ">" // cursor!
		}
		// Render the row
		selectBuilder.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
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
