package k8s_test

import (
	"kubeui/internal/pkg/k8s"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

func TestNewListEventFormat(t *testing.T) {

	comparisonTime := time.Now()
	createdTime := comparisonTime.Add(-(1 * time.Minute))

	tests := []struct {
		event v1.Event
		want  *k8s.ListEventFormat
	}{
		{
			v1.Event{Type: "Warning", Reason: "Some Reason", ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(createdTime)}, ReportingController: "KubeController", Message: "Some Message"},
			&k8s.ListEventFormat{Type: "Warning", Reason: "Some Reason", Age: duration.HumanDuration(comparisonTime.Sub(createdTime)), From: "KubeController", Message: "Some Message"},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := k8s.NewListEventFormat(tt.event, comparisonTime)
			assert.Equal(t, tt.want, got)
		})
	}
}
