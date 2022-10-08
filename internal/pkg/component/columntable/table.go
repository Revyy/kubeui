package columntable

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
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

// Selection represents the act of selecting a row.
type Selection struct {
	Id string
}

// Deletion represents the act of deleting a row.
type Deletion struct {
	Id string
}

// UpdateRowsAndColumns updates the columns and rows of the table.
type UpdateRowsAndColumns struct {
	Columns []*Column
	Rows    []*Row
}

// UpdateHighlighted sets a new previous choice values.
type UpdateHighlighted struct {
	RowId string
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

// Column defined a column of the table.
type Column struct {
	Desc  string
	Width int
}

// Row defined a row of the table.
type Row struct {
	Id     string
	Values []string
}

// Options specifies additional options to be considered when creating a searchtable.
type Options struct {
	// Used to modify help texts for keys.
	SingularItemName string

	// If true, then the search field will be active to start with.
	StartInSearchMode bool
}

type ColumnTable struct {
	keys   *KeyMap
	cursor int

	highlighted string

	allowDelete      bool
	currentRowsSlice []*Row
	currentPage      int
	pageSize         int
	numPages         int
	numRows          int

	columns []*Column
	rows    []*Row

	numFilteredRows int
	searchField     textinput.Model
	searchMode      bool
}

// Returns a list of keybindings to be used in help text.
func (st ColumnTable) KeyList() []key.Binding {
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

func New(columns []*Column, rows []*Row, pageSize int, previousChoice string, allowDelete bool, options Options) ColumnTable {
	searchField := textinput.New()
	searchField.Placeholder = ""
	searchField.Focus()
	searchField.CharLimit = 156
	searchField.Width = 20

	numPages := int(math.Ceil(float64(len(rows)) / float64(pageSize)))
	numRows := len(rows)

	sliceStart, sliceEnd := calcSlice(numRows, 0, pageSize)

	return ColumnTable{
		keys: newKeyMap(options.SingularItemName),

		currentRowsSlice: rows[sliceStart:sliceEnd],
		allowDelete:      allowDelete,
		highlighted:      previousChoice,
		pageSize:         pageSize,
		numPages:         numPages,
		searchField:      searchField,
		numRows:          numRows,
		numFilteredRows:  numRows,
		searchMode:       options.StartInSearchMode,
		columns:          columns,
		rows:             rows,
	}
}

func (ct ColumnTable) Update(msg tea.Msg) (ColumnTable, tea.Cmd) {

	var cmd tea.Cmd

	if ct.searchMode {
		ct, cmd = updateInSearchMode(ct, msg)
	} else {
		ct, cmd = updateInselectMode(ct, msg)
		if cmd != nil {
			return ct, cmd
		}
	}

	switch m := msg.(type) {
	case UpdateRowsAndColumns:
		ct.rows = m.Rows
		ct.columns = m.Columns
	case UpdateHighlighted:
		ct.highlighted = m.RowId
	}

	// Filter rows based on the search value.
	filteredRows := []*Row{}

	for _, row := range ct.rows {
		if strings.Contains(row.Id, ct.searchField.Value()) {
			filteredRows = append(filteredRows, row)
		}
	}

	// If we have a search result that is different than the last result we reset the page.
	if numFilteredItems := len(filteredRows); numFilteredItems != ct.numFilteredRows {
		ct.numFilteredRows = numFilteredItems
		ct.currentPage = 0
		ct.numPages = int(math.Ceil(float64(ct.numFilteredRows) / float64(ct.pageSize)))
	}

	// Calculate which items should be displayed based on the current page and the pageSize.
	sliceStart, sliceEnd := calcSlice(ct.numFilteredRows, ct.currentPage, ct.pageSize)
	ct.currentRowsSlice = filteredRows[sliceStart:sliceEnd]

	// If the selection on the previous page was at a higher index than the current pages total items
	// then we reset it to avoid having a missing cursor.
	if ct.cursor > len(ct.currentRowsSlice)-1 {
		ct.cursor = 0
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return ct, cmd

}

func updateInselectMode(ct ColumnTable, msg tea.Msg) (ColumnTable, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch {

		case key.Matches(msg, ct.keys.Search):
			ct.searchMode = true
			return ct, nil
		// The "up" and "k" keys move the cursor up
		case key.Matches(msg, ct.keys.Up):
			if ct.cursor <= 0 {
				ct.searchMode = true
				return ct, nil
			}
			ct.cursor--

		// The "down" and "j" keys move the cursor down
		case key.Matches(msg, ct.keys.Down):
			if ct.cursor < len(ct.currentRowsSlice)-1 {
				ct.cursor++
			}

		case key.Matches(msg, ct.keys.Left):
			if ct.currentPage > 0 {
				ct.currentPage--
			}

		case key.Matches(msg, ct.keys.Right):
			if ct.currentPage < ct.numPages-1 {
				ct.currentPage++
			}

		case key.Matches(msg, ct.keys.Enter):
			row := ct.currentRowsSlice[ct.cursor]
			return ct, func() tea.Msg {
				return Selection{Id: row.Id}
			}
		case key.Matches(msg, ct.keys.Delete) && ct.allowDelete:
			row := ct.currentRowsSlice[ct.cursor]
			return ct, func() tea.Msg {
				return Deletion{Id: row.Id}
			}
		}
	}

	return ct, nil
}

func updateInSearchMode(ct ColumnTable, msg tea.Msg) (ColumnTable, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, ct.keys.ExitSearch):
			ct.searchMode = false
			return ct, nil
		}
	}
	ct.searchField, _ = ct.searchField.Update(msg)
	return ct, nil
}

func (ct ColumnTable) View() string {

	var mainBuilder strings.Builder

	searchStyle := selectedPageStyle
	if !ct.searchMode {
		searchStyle = unSelectedPageStyle
	}

	mainBuilder.WriteString(searchStyle.Render(ct.searchField.View()) + "\n\n\n")

	var selectBuilder strings.Builder

	columnsData := slices.Map(ct.columns, func(c *Column) string {
		return lipgloss.NewStyle().Width(c.Width + 2).Render(fmt.Sprintf("  %s", c.Desc))
	})
	mainBuilder.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, columnsData...) + "\n\n")

	// Iterate over the rows in the current page and print them out.
	for i, row := range ct.currentRowsSlice {
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if ct.cursor == i {
			cursor = ">" // cursor!
		}

		rowData := []string{}

		for i, value := range row.Values {
			rowData = append(rowData, lipgloss.NewStyle().Width(ct.columns[i].Width+2).Render(value))
		}

		// Render the row
		if row.Id == ct.highlighted {
			selectBuilder.WriteString(fmt.Sprintf("%s %s\n", cursor, highlightedStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, rowData...))))
		} else {
			selectBuilder.WriteString(fmt.Sprintf("%s %s\n", cursor, lipgloss.JoinHorizontal(lipgloss.Left, rowData...)))
		}

	}

	// Start building the pageinator view.
	paginatorView := "\n\n"
	// If we are not at the first page then display a left arrow indicating that we can go left.
	if ct.currentPage > 0 {
		paginatorView += "< "
	} else { // Else just print space to fill the void of the arrow.
		paginatorView += "  "
	}

	// Print out the pages in order as [1 2 3 4] etc.
	paginatorView += "[ "
	for i := 0; i < ct.numPages; i++ {
		if i == ct.currentPage {
			paginatorView += selectedPageStyle.Render(fmt.Sprintf("%d", i+1))
		} else {
			paginatorView += unSelectedPageStyle.Render(fmt.Sprintf("%d", i+1))
		}
		// Add space between numbers inside the brackets.
		if i < ct.numPages-1 {
			paginatorView += "  "
		}
	}
	paginatorView += " ]"

	// If we are not at the last page then display a right arrow indicating that we can go right.
	if ct.currentPage < ct.numPages-1 {
		paginatorView += " >"
	}

	selectBuilder.WriteString(paginatorView)

	selectStyle := selectedPageStyle
	if ct.searchMode {
		selectStyle = unSelectedPageStyle
	}

	mainBuilder.WriteString(selectStyle.Render(selectBuilder.String()))

	return mainBuilder.String()
}

func (n ColumnTable) Init() tea.Cmd {
	return func() tea.Msg { return "" }
}
