package main

import (
	"kubeui/internal/app/kubeui"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/logging"
	"log"
)

var logFilePath = "./logs/kubeui.log"

func main() {
	logger := logging.NewZapLogger(logFilePath)

	logger.Info("Program starting")

	clientSet, err := k8s.NewKClientSet()

	if err != nil {
		log.Fatalf("%v", err)
	}

	model, err := kubeui.NewModel(clientSet, logger)

	if err != nil {
		log.Fatalf("%v", err)
	}

	program := kubeui.NewProgram(model)

	kubeui.StartProgram(program)

}
