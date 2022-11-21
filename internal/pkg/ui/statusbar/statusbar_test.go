package statusbar_test

import (
	"kubeui/internal/pkg/ui/statusbar"
	"testing"
)

var expected = `Test: 10 Test2: value Test3: val
ue2
────────────────────────────────
───`

// This test is mainly here as a regression test.

func TestStatusBar(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		values []string
		want   string
	}{
		{
			"Basic statusbar",
			32,
			[]string{"Test: 10",
				"Test2: value",
				"Test3: value2",
			},
			expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusbar.New(tt.width, " ", tt.values...); got != tt.want {
				t.Errorf("\nStatusBar() = \n%v\n want \n%v", got, tt.want)
			}
		})
	}
}
