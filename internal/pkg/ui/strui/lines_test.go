package strui_test

import (
	"kubeui/internal/pkg/ui/strui"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineBreak(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		maxWidth int
		want     string
	}{
		{"one word that is too long", "abc123", 5, "abc12"},
		{"word shorter than max", "abc1", 5, "abc1"},
		{"split with space", "hello my friend split here", 16, "hello my friend\nsplit here"},
		{"basic split on not space", "hey this will be split in two!", 15, "hey this will \nbe split in two\n!"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strui.LineBreak(tt.str, tt.maxWidth); got != tt.want {
				t.Errorf("LineBreak() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		want   string
	}{
		{"maxlen smaller than 4 should give empty string: 3", "abc123", 3, ""},
		{"maxlen smaller than 4 should give empty string: 2", "abc123", 2, ""},
		{"maxlen smaller than 4 should give empty string: 1", "abc123", 1, ""},
		{"maxlen smaller than 4 should give empty string: 0", "abc123", 0, ""},
		{"maxlen smaller than 4 should give empty string: -1", "abc123", -1, ""},

		{"Should truncate one word that is too long", "abc123", 5, "ab..."},
		{"Should cut off sentence", "a sentence that will be cut off", 13, "a sentence..."},
		{"Should not truncate sentence", "a sentence that will be returned", 300, "a sentence that will be returned"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strui.Truncate(tt.text, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}
