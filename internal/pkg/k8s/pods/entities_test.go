package pods_test

import (
	"kubeui/internal/pkg/k8s/pods"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

type containerNamesTest struct {
	name     string
	pod      *pods.Pod
	expected []string
}

var containerNamesTests = []containerNamesTest{
	{
		name:     "Should work with nil receivers",
		pod:      nil,
		expected: []string{},
	},
	{
		name:     "Should work with default values",
		pod:      &pods.Pod{},
		expected: []string{},
	},
	{
		name:     "Should return the correct container names",
		pod:      &pods.Pod{Pod: v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{{Name: "test"}, {Name: "test2"}}}}},
		expected: []string{"test", "test2"},
	},
}

func TestPod_ContainerNames(t *testing.T) {

	for _, test := range containerNamesTests {
		t.Run(test.name, func(t *testing.T) {
			got := test.pod.ContainerNames()
			assert.Equal(t, got, test.expected, test.name)
		})
	}
}
