package k8s_test

import (
	"kubeui/internal/pkg/k8s"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

type ContainerNamesTest struct {
	name     string
	pod      *k8s.Pod
	expected []string
}

var ContainerNamesTests = []ContainerNamesTest{
	{
		name:     "Should work with nil receivers",
		pod:      nil,
		expected: []string{},
	},
	{
		name:     "Should work with default values",
		pod:      &k8s.Pod{},
		expected: []string{},
	},
	{
		name:     "Should return the correct container names",
		pod:      &k8s.Pod{Pod: v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{{Name: "test"}, {Name: "test2"}}}}},
		expected: []string{"test", "test2"},
	},
}

func TestPod_ContainerNames(t *testing.T) {

	for _, test := range ContainerNamesTests {
		t.Run(test.name, func(t *testing.T) {
			got := test.pod.ContainerNames()
			assert.Equal(t, got, test.expected, test.name)
		})
	}
}
