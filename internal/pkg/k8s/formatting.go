package k8s

import (
	"fmt"
	"time"

	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// ListPodFormat contains information about a pod, as shown when running `kubectl get pods`.
type ListPodFormat struct {
	Name     string
	Ready    string
	Status   string
	Restarts string
	Age      string
}

// NewListPodFormat collects the ListPodFormat information for a given pod.
func NewListPodFormat(pod v1.Pod) *ListPodFormat {

	podStatus := PodStatus(pod)
	lastRestart := PodLastRestarted(pod)
	restartCount := PodRestartCount(pod)
	readyCount := ContainerReadyCount(pod)

	lastRestartStr := ""
	if !lastRestart.IsZero() {
		lastRestartStr = fmt.Sprintf(" (%s ago)", duration.HumanDuration(time.Since(lastRestart)))
	}

	podAgeStr := PodAge(pod, time.Now())
	podRestartStr := fmt.Sprintf("%d%s", restartCount, lastRestartStr)
	containersReadyStr := fmt.Sprintf("%d/%d", readyCount, len(pod.Status.ContainerStatuses))

	return &ListPodFormat{
		Name:     pod.Name,
		Ready:    containersReadyStr,
		Status:   podStatus,
		Restarts: podRestartStr,
		Age:      podAgeStr,
	}
}

// PodLastRestarted returns the time of when the pod was last restarted.
func PodLastRestarted(pod v1.Pod) time.Time {

	lastRestart := metav1.Time{}

	for _, container := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		// Update the last restarted time.
		if container.LastTerminationState.Terminated != nil {
			if lastRestart.Before(&container.LastTerminationState.Terminated.FinishedAt) {
				lastRestart = container.LastTerminationState.Terminated.FinishedAt
			}
		}
	}

	return lastRestart.Time
}

// ContainerReadyCount counts the number of containers in a pod who are ready.
func ContainerReadyCount(pod v1.Pod) int {
	readyCount := 0

	for _, container := range pod.Status.ContainerStatuses {
		if container.Ready && container.State.Running != nil {
			readyCount++
		}
	}

	return readyCount
}

// PodRestartCount calculates the number of times the pod has been restarted.
func PodRestartCount(pod v1.Pod) int {
	restartCount := 0

	for _, container := range pod.Status.ContainerStatuses {
		restartCount += int(container.RestartCount)
	}

	return restartCount
}

// PodAge calculates the age of the pod, based on its creation timestamp.
// The result is a string created by using the standard library function duration.HumanDuration.
func PodAge(pod v1.Pod, comparisonTime time.Time) string {
	return duration.HumanDuration(comparisonTime.Sub(pod.CreationTimestamp.Time))
}

// ListEventFormat contains information about an event, as shown when running `kubectl describe pod` for example.
type ListEventFormat struct {
	Type    string
	Reason  string
	Age     string
	From    string
	Message string
}

// NewListEventFormat collects the ListEventFormat information.
func NewListEventFormat(event v1.Event, now time.Time) *ListEventFormat {

	return &ListEventFormat{
		Type:    event.Type,
		Reason:  event.Reason,
		Age:     duration.HumanDuration(now.Sub(event.CreationTimestamp.Time)),
		From:    event.ReportingController,
		Message: event.Message,
	}
}

func PodStatus(pod v1.Pod) string {
	initializing := IsPodInitializing(pod)

	// We first collect information from the initContainers in the pod.
	// Depending on the status of those we are in the initializing/creation phase of a pod or the running phase of a pod.
	var status string

	if initializing {
		status = PodInitStatus(pod)
	} else {
		status = PodNonInitStatus(pod)
	}

	// If the pod is scheduled for deletion.
	if pod.DeletionTimestamp != nil {
		status = "Terminating"
	}

	// If it is scheduled for deletion with reason NodeLost.
	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		status = "Unknown"
	}

	return status
}

func IsPodInitializing(pod v1.Pod) bool {
	initializing := false

	for _, initContainer := range pod.Status.InitContainerStatuses {
		// If the initcontainer is terminated with a zero exitcode then we can assume that we are done
		// initializing.
		switch {
		case initContainer.State.Terminated != nil && initContainer.State.Terminated.ExitCode == 0:
			continue
		default:
			initializing = true
		}
	}

	return initializing
}

// PodInitStatus returns the status of a pod assuming it is in the initializing phase.
// Meaning that is has failed or waiting init containers.
func PodInitStatus(pod v1.Pod) string {
	status := defaultPodStatus(pod)

	for i, initContainer := range pod.Status.InitContainerStatuses {

		// If the initcontainer is terminated with a zero exitcode then we can assume that we are done
		// initializing.
		switch {
		case initContainer.State.Terminated != nil && initContainer.State.Terminated.ExitCode == 0:
			continue
		// If the initcontainer is terminated for some other reason then we can assume that the pod will restart,
		// Which means we are still initializing the pod.
		case initContainer.State.Terminated != nil:
			status = initContainerTerminationReason(initContainer)
		// If we are waiting but not with the reason PodInitializing then we are still initializing.
		case initContainer.State.Waiting != nil && notEmptyOr(initContainer.State.Waiting.Reason, "PodInitializing"):
			status = "Init:" + initContainer.State.Waiting.Reason

		default:
			status = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
		}
	}

	return status
}

// initContainerTerminationReason returns the termination reason for an initContainer.
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

// PodNonInitStatus returns the status of a pod assuming it is done initializing.
func PodNonInitStatus(pod v1.Pod) string {

	status := defaultPodStatus(pod)
	hasRunningContainer := false

	for _, container := range slices.Reverse(pod.Status.ContainerStatuses) {

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

	return status
}

// containerTerminationReason returns the termination reason for a regular container.
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

// function that checks if a pod is ready based on its conditions
func podReadyCondition(conditions []v1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// defaultPodStatus fetches the status of a pod based on its phase and potential status reason.
func defaultPodStatus(pod v1.Pod) string {
	// Default of pod
	status := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		status = pod.Status.Reason
	}

	return status
}

// notEmptyOr returns true if the string is not the empty string and it is not equal to notStr.
func notEmptyOr(str, notStr string) bool {
	return len(str) > 0 && str != notStr
}
