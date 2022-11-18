package kubeui_test

import (
	"kubeui/internal/pkg/kubeui"
	"testing"
)

func TestStatusBar(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		values []string
		want   string
	}{
		{
			"Basic statusbar",
			30,
			[]string{"Test: 10",
				"Test2: value",
				"Test3: value2",
			},
			"Test: 10 Test2: value Test3:  \nvalue2                        \n──────────────────────────────",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kubeui.StatusBar(tt.width, " ", tt.values...); got != tt.want {
				t.Errorf("\nStatusBar() = \n%v\n want \n%v", got, tt.want)
			}
		})
	}
}
