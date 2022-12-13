package main

import (
	"kubeui/internal/app/cxs"
	"kubeui/internal/app/pods"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/k8s/namespace"
	k8spods "kubeui/internal/pkg/k8s/pods"
	"kubeui/internal/pkg/kubeui"
	"log"

	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"
)

type args struct {
	Program    string `arg:"positional" help:"Subcommand to run, one of [cxs, pods]"`
	KubeConfig string `arg:"-c" help:"Absolute path to the kubeconfig file"`
}

func main() {

	// Parse arguments given the 'args' struct
	args := &args{}
	arg.MustParse(args)

	// If a specific kubeconfig file is specified then we load that, otherwise the defaults will be loaded.

	clientConfig := k8s.NewClientConfig("", args.KubeConfig)

	rawConfig, err := clientConfig.RawConfig()

	if err != nil {
		log.Fatalf("failed to load config")
	}

	configAccess := clientConfig.ConfigAccess()

	clientSet, err := k8s.NewKClientSet(args.KubeConfig, configAccess)

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var m tea.Model

	switch args.Program {
	case "cxs":
		m = cxs.NewModel(k8s.NewContextClientImpl(configAccess, rawConfig, nil))
	case "pods":
		m = pods.NewModel(k8s.NewContextClientImpl(configAccess, rawConfig, nil), k8s.NewK8sService(
			k8spods.NewRepository(clientSet.CoreV1()),
			namespace.NewRepository(clientSet.CoreV1()),
		))
	default:
		log.Fatalf("no command called %s", args.Program)
	}

	program := kubeui.NewProgram(m, true)
	kubeui.StartProgram(program)

}
