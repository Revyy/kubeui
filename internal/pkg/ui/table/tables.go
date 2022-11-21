package table

import (
	"sort"

	"github.com/life4/genesis/maps"

	"k8s.io/utils/integer"
)

// StringMapColumnsAndRows creates the neccessary columns and rows in order to display a map[string]string as a table.
func StringMapColumnsAndRows(maxWidth int, col1 string, col2 string, data map[string]string) ([]*DataColumn, []*DataRow) {

	columns := []*DataColumn{
		{Desc: col1, Width: len(col1)},
		{Desc: col2, Width: len(col2)},
	}

	keys := maps.Keys(data)
	sort.Strings(keys)

	rows := []*DataRow{}

	for _, key := range keys {

		value := data[key]
		// Update widths of the columns
		columns[0].Width = integer.IntMax(columns[0].Width, len(key))

		remainingWidth := maxWidth - columns[0].Width
		columns[1].Width = integer.IntMax(remainingWidth-1, len(value))

		rows = append(rows, &DataRow{
			Values: []string{key, value},
		})
	}

	return columns, rows
}
