package table_test

import (
	"kubeui/internal/pkg/ui/table"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumnsString(t *testing.T) {

	tests := []struct {
		name    string
		columns []table.DataColumn
		want    string
	}{
		{"Empty columns slice should give empty string", []table.DataColumn{}, ""},
		{"Single column with matching width should give back the column description", []table.DataColumn{{"Column 1", 8}}, "Column 1"},
		{"Columns should be rendered according to their width, filling out with spaces", []table.DataColumn{{"Column 1", 10}, {"Column 2", 10}, {"Column Last", 11}}, "Column 1  Column 2  Column Last"},
		{"Column width smaller than length of desc should be truncated", []table.DataColumn{{"Column 1", 10}, {"Column 2", 6}, {"Column Last", 11}}, "Column 1  ColumnColumn Last"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := table.ColumnsToString(tt.columns)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRowsString(t *testing.T) {

	tests := []struct {
		name    string
		columns []table.DataColumn
		rows    []table.DataRow
		want    string
	}{
		{"Empty columns should result in empty string", []table.DataColumn{}, []table.DataRow{{Values: []string{"Value 1"}}}, ""},
		{"Empty rows should result in empty string", []table.DataColumn{{Desc: "Column 1", Width: 10}}, []table.DataRow{}, ""},
		{
			"Rows with a number of value not matching the number of columns should be skipped",
			[]table.DataColumn{
				{Desc: "Column 1", Width: 10},
				{Desc: "Column 2", Width: 10},
			},
			[]table.DataRow{
				{Values: []string{"Value 1", "Value 2"}},
				{Values: []string{"Value 3"}},
				{Values: []string{"Value 4", "Value 5"}},
			},
			"Value 1   Value 2   \nValue 4   Value 5   ",
		},
		{
			"Rows should be rendered ",
			[]table.DataColumn{
				{Desc: "Column 1", Width: 10},
				{Desc: "Column 2", Width: 10},
				{Desc: "Column 3", Width: 8},
			},
			[]table.DataRow{
				{Values: []string{"Value 1", "Value 2", "Value 3"}},
				{Values: []string{"Value 4", "Value 5", "Value 6"}},
				{Values: []string{"Value 7", "Value 8", "Value 9"}},
			},
			"Value 1   Value 2   Value 3 \nValue 4   Value 5   Value 6 \nValue 7   Value 8   Value 9 ",
		},
		{
			"Values which are too long should be split on multiple rows",
			[]table.DataColumn{
				{Desc: "Column 1", Width: 10},
				{Desc: "Column 2", Width: 10},
				{Desc: "Column 3", Width: 8},
			},
			[]table.DataRow{
				{Values: []string{"Value 1", "Value 222222", "Value 3"}},
				{Values: []string{"Value 4", "Value 5", "Value 6"}},
			},
			"Value 1   Value     Value 3 \n          222222            \nValue 4   Value 5   Value 6 ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := table.RowsToString(tt.columns, tt.rows)
			assert.Equal(t, tt.want, got)
		})
	}
}
