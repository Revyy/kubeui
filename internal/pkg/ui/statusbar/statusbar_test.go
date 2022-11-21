package statusbar_test

import (
	"kubeui/internal/pkg/ui/statusbar"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expectedWithLineBreak = `Test: 10 Test2: value Test3: val
ue2
────────────────────────────────
───`

func TestStatusBar(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		values []string
		want   string
	}{
		{"Width 0 should give empty string", 0, []string{"Some key"}, ""},
		{"Basic statusbar", 100, []string{"Test: 10", "Test2: value", "Test3: value2"}, "Test: 10 Test2: value Test3: value2\n───────────────────────────────────"},
		{"Basic statusbar with linebreak due to truncation", 32, []string{"Test: 10", "Test2: value", "Test3: value2"}, expectedWithLineBreak},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusbar.New(tt.width, " ", tt.values...)
			assert.Equal(t, tt.want, got)
		})
	}
}
