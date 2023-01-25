package kubeui

// Context contains the context of the kubeui application.
type Context struct {

	// Currently selected namespace
	Namespace string

	// Name of currently selected pod.
	SelectedPod string
}
