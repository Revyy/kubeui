package selection_test

import (
	"kubeui/internal/pkg/ui/selection"
	"testing"

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
		{"cursor missmatch", args{2, 100, []string{"Header 1", "Header 2"}}, "Header 1 Header 2"},
		// This is here as a regression test as it is really hard to test these ansi-strings in a good way by specifying the expected values.
		{"cursor 1 should generate underline under first header and color code the text", args{0, 100, []string{"Header 1", "Header 2"}}, "\x1b[4;4mH\x1b[0m\x1b[4;4me\x1b[0m\x1b[4;4ma\x1b[0m\x1b[4;4md\x1b[0m\x1b[4;4me\x1b[0m\x1b[4;4mr\x1b[0m\x1b[4m \x1b[0m\x1b[4;4m1\x1b[0m Header 2"},
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
		{"should generate correct string", args{[]string{"Item 1", "Item 2", "Item 3"}, "Item 1", 100}, "[1] Item 1 [2] Item 2 [3] Item 3"},
		{"should wrap string correctly", args{[]string{"Item 1", "Item 2", "Item 3"}, "Item 1", 21}, "[1] Item 1 [2] Item 2\n[3] Item 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selection.HorizontalList(tt.args.items, tt.args.selectedItem, tt.args.maxWidth)
			assert.Equal(t, tt.want, got)
		})
	}
}
