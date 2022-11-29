package k8s_test

import (
	"fmt"
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

func TestPodAge(t *testing.T) {

	comparisonTime := time.Now()

	tests := []struct {
		pod  v1.Pod
		want string
	}{
		{v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(comparisonTime.Add(-1 * time.Second))}}, "1s"},
		{v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(comparisonTime.Add(-2 * time.Minute))}}, "2m"},
		{v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(comparisonTime.Add(-2*time.Hour + -3*time.Minute))}}, "123m"},
		{v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(comparisonTime.Add(-49 * time.Hour))}}, "2d1h"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("TestPodAge %d", i), func(t *testing.T) {
			got := k8s.PodAge(tt.pod, comparisonTime)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPodRestartCount(t *testing.T) {

	duplicateContainerStatus := func(s v1.ContainerStatus, count int, increasing bool) []v1.ContainerStatus {
		containerStatus := []v1.ContainerStatus{}

		for i := 0; i < count; i++ {

			st := s

			if increasing {
				st.RestartCount = int32(i + 1)
			}

			containerStatus = append(containerStatus, st)
		}

		return containerStatus
	}

	tests := []struct {
		name string
		pod  v1.Pod
		want int
	}{
		{"No containers", v1.Pod{}, 0},
		{"Containers without restart", v1.Pod{Status: v1.PodStatus{ContainerStatuses: duplicateContainerStatus(v1.ContainerStatus{RestartCount: 0}, 10, false)}}, 0},
		{"Containers with same restart counts", v1.Pod{Status: v1.PodStatus{ContainerStatuses: duplicateContainerStatus(v1.ContainerStatus{RestartCount: 3}, 10, false)}}, 30},
		{"Containers with differest restart counts", v1.Pod{Status: v1.PodStatus{ContainerStatuses: duplicateContainerStatus(v1.ContainerStatus{RestartCount: 0}, 10, true)}}, 55},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s.PodRestartCount(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerReadyCount(t *testing.T) {
	duplicateContainerStatus := func(s v1.ContainerStatus, count int) []v1.ContainerStatus {
		containerStatus := []v1.ContainerStatus{}

		for i := 0; i < count; i++ {
			st := s
			containerStatus = append(containerStatus, st)
		}

		return containerStatus
	}

	tests := []struct {
		name string
		pod  v1.Pod
		want int
	}{
		{"No containers", v1.Pod{}, 0},

		// Containers are considered running if the have ReadyStatus true and their Running state is set.
		{"10 Running containers", v1.Pod{Status: v1.PodStatus{ContainerStatuses: duplicateContainerStatus(v1.ContainerStatus{Ready: true, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}}, 10)}}, 10},

		{"5 running, 10 not running", v1.Pod{Status: v1.PodStatus{
			ContainerStatuses: append(
				duplicateContainerStatus(v1.ContainerStatus{Ready: true, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}}, 5),
				append(
					duplicateContainerStatus(v1.ContainerStatus{Ready: false, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}}, 5),
					duplicateContainerStatus(v1.ContainerStatus{Ready: true, State: v1.ContainerState{Running: nil}}, 5)...,
				)...,
			),
		}}, 5},
		{"10 not running", v1.Pod{Status: v1.PodStatus{ContainerStatuses: duplicateContainerStatus(v1.ContainerStatus{Ready: false, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}}, 10)}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s.ContainerReadyCount(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPodLastRestarted(t *testing.T) {

	referenceTime := time.Now()

	tests := []struct {
		name string
		pod  v1.Pod
		want time.Time
	}{
		{"No restarts should give default time value", v1.Pod{}, time.Time{}},
		{"It should calculate LastRestarted using init containers", v1.Pod{Status: v1.PodStatus{InitContainerStatuses: []v1.ContainerStatus{
			{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime)}}},
		}}}, referenceTime},

		{"It should calculate LastRestarted using both init containers and regular containers, picking the latest one", v1.Pod{Status: v1.PodStatus{
			InitContainerStatuses: []v1.ContainerStatus{
				{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime)}}},
			},
			ContainerStatuses: []v1.ContainerStatus{
				{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime.Add(3 * time.Minute))}}},
			},
		}}, referenceTime.Add(3 * time.Minute)},

		{"It should pick the latest restart time and not just the latest container", v1.Pod{Status: v1.PodStatus{
			InitContainerStatuses: []v1.ContainerStatus{
				{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime)}}},
			},
			ContainerStatuses: []v1.ContainerStatus{
				{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime.Add(5 * time.Minute))}}},
				{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime.Add(3 * time.Minute))}}},
			},
		}}, referenceTime.Add(5 * time.Minute)},

		{"It should ignore containers without a terminated state", v1.Pod{Status: v1.PodStatus{InitContainerStatuses: []v1.ContainerStatus{
			{LastTerminationState: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(referenceTime)}}},
			{LastTerminationState: v1.ContainerState{Terminated: nil}},
		}}}, referenceTime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s.PodLastRestarted(tt.pod)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsPodInitializing(t *testing.T) {

	tests := []struct {
		name string
		pod  v1.Pod
		want bool
	}{
		{"Pod is not initializing if no init containers exist", v1.Pod{}, false},
		{"If all init containers have terminated with exit code 0 then we are not initializing", v1.Pod{Status: v1.PodStatus{InitContainerStatuses: []v1.ContainerStatus{
			{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
			{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
		}}}, false},
		{"If at least one init container have not terminated with exit code 0 then we are initializing", v1.Pod{Status: v1.PodStatus{InitContainerStatuses: []v1.ContainerStatus{
			{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
			{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: ""}}},
		}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s.IsPodInitializing(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPodInitStatus(t *testing.T) {
	tests := []struct {
		name string
		pod  v1.Pod
		want string
	}{
		{"With no init containers and no status reason we should get the pod phase", v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, string(v1.PodRunning)},

		{"With no init containers then we should get the status reason if it exists", v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning, Reason: "Reason"}}, "Reason"},

		{"If all init containers have terminated successfully then we should get the default status", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
				},
			},
		}, "Reason"},

		{"We should get the termination reason for the last init container if it exists", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Reason: "Reason 2"}}},
				},
			},
		}, "Init: Reason 2"},

		{"We should get the signal code for the last init container if exists and no reason exists", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Signal: 10}}},
				},
			},
		}, "Init:Signal:10"},

		{"We should get the exit code for the last init container if exists and no reason or signal exists", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1}}},
				},
			},
		}, "Init:ExitCode:1"},

		{"Exit reason should be prioritized over singal and exit code", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Reason: "Reason 2", Signal: 10}}},
				},
			},
		}, "Init: Reason 2"},

		{"Exit signal should be prioritized over exit code", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Signal: 10}}},
				},
			},
		}, "Init:Signal:10"},

		// Does not apply to the special reason PodInitializing or the empty reason(empty string).
		{"If the last init container has status waiting with a reason then we should get the reason", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "WaitReason"}}},
				},
			},
		}, "Init:WaitReason"},

		// Init containers are executed in order according to the initContainers field of the PodSpec.
		// For the status, we only care about the number of init containers, not what they contain.
		// A container status is only added to the list of InitContainerStatuses once it has started executing.
		// So we can rely on that the last InitContainerStatus belongs to the currently executing container or the last terminated one.
		{"If none of the above applies then we return how many init containers have executed", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{}},
				},
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{},
					{},
					{},
				},
			},
		}, "Init:1/3"},

		// In this case we should expect the init process to be completed soon.
		{"If the last container status has state wating with reason PodInitializing then we also show many init containers have executed", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "PodInitializing"}}},
				},
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{},
					{},
				},
			},
		}, "Init:1/2"},

		// In this case we do not yet have a reason for the waiting status of the last init container so we show a generic status while waiting.
		{"If the last container status has state wating with an empty reason then we also show many init containers have executed", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				InitContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: ""}}},
				},
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{},
					{},
				},
			},
		}, "Init:1/2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s.PodInitStatus(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPodNonInitStatus(t *testing.T) {
	tests := []struct {
		name string
		pod  v1.Pod
		want string
	}{
		{"With no container statuses and no status reason we should get the pod phase", v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning}}, string(v1.PodRunning)},

		{"With no container statuses then we should get the status reason if it exists", v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning, Reason: "Reason"}}, "Reason"},

		{"If the first container has status waiting with a reason then we should get the reason", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "WaitReason"}}},
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "WaitReason2"}}},
				},
			},
		}, "WaitReason"},

		{"If the first container is terminated and we have a reason", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "WaitReason"}}},
					{State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "WaitReason2"}}},
				},
			},
		}, "WaitReason"},

		{"We should get the termination reason for the first container if it exists", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Reason: "Reason 1"}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
				},
			},
		}, "Reason 1"},

		{"We should get the signal code for the first container if exists and no reason exists", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{Signal: 10}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0, Reason: "Reason 2"}}},
				},
			},
		}, "Signal:10"},

		{"We should get the exit code for the first container if exists and no reason or signal exists", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0, Reason: "Reason 2"}}},
				},
			},
		}, "ExitCode:1"},

		{"Exit reason should be prioritized over singal and exit code", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Reason: "Reason 2", Signal: 10}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
				},
			},
		}, "Reason 2"},

		{"Exit signal should be prioritized over exit code", v1.Pod{
			Status: v1.PodStatus{
				Phase:  v1.PodRunning,
				Reason: "Reason",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1, Signal: 10}}},
					{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
				},
			},
		}, "Signal:10"},

		// This is a tricky one.
		// Basically if we still have a running container, no terminated ones, but one that is completed.
		// Then we check if the pod has a ready condition that is true, if it is then we are running and
		// the completed container probably ran as part of some startup process.
		{"One running container, one completed and pod ready", v1.Pod{
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{Type: v1.PodReady, Status: v1.ConditionTrue},
				},
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{Ready: true, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}},
					{Ready: true, State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "Completed"}}},
				},
			},
		}, "Running"},

		// This is a tricky one.
		// Basically if we still have a running container, no terminated ones, but one that is completed.
		// Then we check if the pod has a ready condition that is true, if it is not then we are not ready yet.
		// This is probably a temporary waiting status between a startup container running and the pod receiving the ready condition.
		{"One running container, one completed and pod not ready", v1.Pod{
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{Type: v1.PodReady, Status: v1.ConditionFalse},
				},
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{Ready: true, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}},
					{Ready: true, State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "Completed"}}},
				},
			},
		}, "NotReady"},

		// Same conditions as the ones above but in this case the pod is waiting with an actual reason.
		// So in this case that reason should be used instead.
		{"One running container, one waiting and pod ready", v1.Pod{
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{Type: v1.PodReady, Status: v1.ConditionTrue},
				},
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{Ready: true, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}},
					{Ready: true, State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "Some Reason"}}},
				},
			},
		}, "Some Reason"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s.PodNonInitStatus(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
