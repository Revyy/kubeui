package selection_test

import (
	"testing"

	"kubeui/internal/pkg/ui/selection"

	"github.com/stretchr/testify/assert"
)

func TestTabs(t *testing.T) {
	type args struct {
		cursor   int
		maxWidth int
		headers  []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"maxWidth 0 should give empty string", args{0, 0, []string{"Header 1", "Header 2"}}, ""},
		// Cursor is outside the max index of the headers slice.
		{"cursor missmatch", args{2, 20, []string{"Header 1", "Header 2"}}, "Header 1 Header 2   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selection.Tabs(tt.args.cursor, tt.args.maxWidth, tt.args.headers)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHorizontalList(t *testing.T) {
	type args struct {
		items        []string
		selectedItem string
		maxWidth     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"maxWidth 0 should give empty string", args{[]string{"Item 1", "Item 2"}, "Item 1", 0}, ""},
		{"no items should give empty string", args{[]string{}, "Item 1", 100}, ""},
		{"should generate correct string", args{[]string{"Item 1", "Item 2", "Item 3"}, "Item 1", 32}, "[1] Item 1 [2] Item 2 [3] Item 3"},
		{"should wrap string correctly", args{[]string{"Item 1", "Item 2", "Item 3"}, "Item 1", 21}, "[1] Item 1 [2] Item 2\n[3] Item 3           "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selection.HorizontalList(tt.args.items, tt.args.selectedItem, tt.args.maxWidth)
			assert.Equal(t, tt.want, got)
		})
	}
}
