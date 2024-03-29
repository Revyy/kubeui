# KubeUI

DISCLAIMER: THIS IS A WORK IN PROGRESS AND EXPERIMENTAL REPOSITORY

A collection of terminal gui based utilities making working with kubernetes easier.

## Installation

First make sure that you have set your GOPATH and that $GOPATH/bin is in your PATH.

Then run:

```bash
make install
```

A binary called kubeui will be installed.

## Contribution
Contribution is welcomed but keep in mind that this is very early days and the code structure is still very much experimental.

### Run tests

Tests are severely lacking at the moment but you can run them using `go test ./...` or just use the make target.

```
make test
```

## Intructions

Using kubeui is as simple as calling `kubeui [PROGRAM]`, where program is one of the programs listed below.

### cxs [STABLE]

A context selection and deletion tool.
Allows you to select a kubecontext and/or deleting a context and identically named cluster and user entries from the the kubeconfig.

### pods [EXPERIMENTAL]
A pod information tool
Allows you to list pods for a selected namespace, with pagination and searching capabilities.

Additional features:

* Deleting a pod
* Inspecting a pod including viewing events and the latest log entries for each container.
