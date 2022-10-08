package k8s

import (
	"fmt"
	"time"

	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/integer"
)

type ListPodFormat struct {
	Name     string
	Ready    string
	Status   string
	Restarts string
	Age      string
}

func NewListPodFormat(pod v1.Pod) *ListPodFormat {
	readyCount := calculateReadyCount(pod.Status.ContainerStatuses)
	var maxRestarts int
	initializing, status, lastRestart := isInitializing(pod)

	if !initializing {
		status, lastRestart, maxRestarts = notInitializing(lastRestart, status, pod)
	}

	// If the pod is scheduled for deletion.
	if pod.DeletionTimestamp != nil {
		status = "Terminating"
	}

	// If it is scheduled for deletion with reason NodeLost.
	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		status = "Unknown"
	}

	lastRestartStr := ""
	if !lastRestart.IsZero() {
		lastRestartStr = fmt.Sprintf(" (%s ago)", duration.HumanDuration(time.Since(lastRestart.Time)))
	}
	containersReadyStr := fmt.Sprintf("%d/%d", readyCount, len(pod.Status.ContainerStatuses))
	podAgeStr := duration.HumanDuration(time.Since(pod.CreationTimestamp.Time))
	podRestartStr := fmt.Sprintf("%d%s", maxRestarts, lastRestartStr)

	return &ListPodFormat{
		Name:     pod.Name,
		Ready:    containersReadyStr,
		Status:   status,
		Restarts: podRestartStr,
		Age:      podAgeStr,
	}
}

func calculateReadyCount(containers []v1.ContainerStatus) int {
	readyCount := 0

	for _, container := range containers {
		if container.Ready && container.State.Running != nil {
			readyCount++
		}
	}

	return readyCount
}

func defaultPodStatus(pod v1.Pod) string {
	// Default of pod
	status := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		status = pod.Status.Reason
	}

	return status
}

func notEmptyOr(str, notStr string) bool {
	return len(str) > 0 && str != notStr
}

func initContainerTerminationReason(container v1.ContainerStatus) string {

	if container.State.Terminated == nil {
		return ""
	}

	if len(container.State.Terminated.Reason) > 0 {
		return "Init: " + container.State.Terminated.Reason
	}

	if container.State.Terminated.Signal != 0 {
		return fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
	}

	return fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
}

func containerTerminationReason(container v1.ContainerStatus) string {

	if container.State.Terminated == nil {
		return ""
	}

	if len(container.State.Terminated.Reason) > 0 {
		return container.State.Terminated.Reason
	}

	if container.State.Terminated.Signal != 0 {
		return fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
	}

	return fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)

}

func isInitializing(pod v1.Pod) (bool, string, metav1.Time) {
	initializing := false
	status := defaultPodStatus(pod)
	lastRestart := metav1.Time{}

	for i, initContainer := range pod.Status.InitContainerStatuses {

		// Update the last restarted time.
		if initContainer.LastTerminationState.Terminated != nil {
			if lastRestart.Before(&initContainer.LastTerminationState.Terminated.FinishedAt) {
				lastRestart = initContainer.LastTerminationState.Terminated.FinishedAt
			}
		}

		// If the initcontainer is terminated with a zero exitcode then we can assume that we are done
		// initializing.
		switch {
		case initContainer.State.Terminated != nil && initContainer.State.Terminated.ExitCode == 0:
			continue
		// If the initcontainer is terminated for some other reason then we can assume that the pod will restart,
		// Which means we are still initializing the pod.
		case initContainer.State.Terminated != nil:
			status = initContainerTerminationReason(initContainer)
			initializing = true
		// If we are waiting but not with the reason PodInitializing then we are still initializing.
		case initContainer.State.Waiting != nil && notEmptyOr(initContainer.State.Waiting.Reason, "PodInitializing"):
			status = "Init:" + initContainer.State.Waiting.Reason
			initializing = true

		default:
			status = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
	}

	return initializing, status, lastRestart
}

func podReadyCondition(conditions []v1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func notInitializing(lastRestart metav1.Time, status string, pod v1.Pod) (string, metav1.Time, int) {
	maxRestarts := 0

	var hasRunningContainer bool

	for _, container := range slices.Reverse(pod.Status.ContainerStatuses) {
		maxRestarts = integer.IntMax(maxRestarts, int(container.RestartCount))

		if container.LastTerminationState.Terminated != nil {
			if lastRestart.Before(&container.LastTerminationState.Terminated.FinishedAt) {
				lastRestart = container.LastTerminationState.Terminated.FinishedAt
			}
		}

		// If we are waiting for the container to start and have a reason.
		if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
			status = container.State.Waiting.Reason
			continue
		}

		if container.State.Terminated != nil {
			status = containerTerminationReason(container)
			continue
		}

		if container.Ready && container.State.Running != nil {
			hasRunningContainer = true
		}
	}

	if status == "Completed" && hasRunningContainer {
		if podReadyCondition(pod.Status.Conditions) {
			status = "Running"
		} else {
			status = "NotReady"
		}
	}

	return status, lastRestart, maxRestarts

}
