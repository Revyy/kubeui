# KubeUI

DISCLAIMER: THIS IS A WORK IN PROGRESS AND EXPERIMENTAL REPOSITORY

A collection of terminal gui based utilities making working with kubernetes easier.

## Installation

First make sure that you have set your GOPATH and that $GOPATH/bin is in your PATH.

Then run:

```bash
make install
```


## Utilities

### cxs

A context selection and deletion tool.
Allows you to select a kubecontext and/or deleting a context and identically named cluster and user entries from the the kubeconfig.
