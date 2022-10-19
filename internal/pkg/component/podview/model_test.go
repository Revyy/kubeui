package podview_test

import (
	"kubeui/internal/pkg/component/podview"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func getModel() podview.Model {
	return podview.New(v1.Pod{}, 100)
}

func TestModel_Update(t *testing.T) {
	// KEYS
	t.Run("Test keystrokes", func(t *testing.T) {
		//model := podview.New(v1.Pod{}, 100)

	})
}
