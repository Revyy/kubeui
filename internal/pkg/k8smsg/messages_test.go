package k8smsg_test

import (
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/k8smsg"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewListNamespacesMsg(t *testing.T) {

	expected := &v1.NamespaceList{Items: []v1.Namespace{{Status: v1.NamespaceStatus{Conditions: []v1.NamespaceCondition{{Message: "test"}}}}}}

	tests := []struct {
		name          string
		namespaceList *v1.NamespaceList
		want          k8smsg.ListNamespacesMsg
	}{
		{"should work with nil", nil, k8smsg.ListNamespacesMsg{NamespaceList: nil}},
		{"should assign the same object", expected, k8smsg.ListNamespacesMsg{NamespaceList: expected}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8smsg.NewListNamespacesMsg(tt.namespaceList)
			assert.Equal(t, tt.want, got, "")
		})
	}
}

func TestNewListPodsMsg(t *testing.T) {

	expected := &v1.PodList{Items: []v1.Pod{{Status: v1.PodStatus{Message: "test"}}}}

	tests := []struct {
		name    string
		podList *v1.PodList
		want    k8smsg.ListPodsMsg
	}{
		{"should work with nil", nil, k8smsg.ListPodsMsg{PodList: nil}},
		{"should assign the same object", expected, k8smsg.ListPodsMsg{PodList: expected}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8smsg.NewListPodsMsg(tt.podList)
			assert.Equal(t, tt.want, got, "")
		})
	}
}

func TestNewPodDeletedMsg(t *testing.T) {

	expected := "pod name"

	tests := []struct {
		name    string
		podName string
		want    k8smsg.PodDeletedMsg
	}{
		{"should work with empty string", "", k8smsg.PodDeletedMsg{Name: ""}},
		{"should assign the same string", expected, k8smsg.PodDeletedMsg{Name: expected}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8smsg.NewPodDeletedMsg(tt.podName)
			assert.Equal(t, tt.want, got, "")
		})
	}
}

func TestNewGetPodMsg(t *testing.T) {

	expected := &k8s.Pod{Pod: v1.Pod{Status: v1.PodStatus{Message: "test"}}}

	tests := []struct {
		name string
		pod  *k8s.Pod
		want k8smsg.GetPodMsg
	}{
		{"should work with nil", nil, k8smsg.GetPodMsg{Pod: nil}},
		{"should assign the same object", expected, k8smsg.GetPodMsg{Pod: expected}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8smsg.NewGetPodMsg(tt.pod)
			assert.Equal(t, tt.want, got, "")
		})
	}
}
