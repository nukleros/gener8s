[![Go Reference](https://pkg.go.dev/badge/github.com/vmware-tanzu-labs/object-code-generator-for-k8s.svg)](https://pkg.go.dev/github.com/vmware-tanzu-labs/object-code-generator-for-k8s)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vmware-tanzu-labs/object-code-generator-for-k8s)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/vmware-tanzu-labs/object-code-generator-for-k8s)](https://goreportcard.com/report/github.com/vmware-tanzu-labs/object-code-generator-for-k8s)
[![GitHub](https://img.shields.io/github/license/vmware-tanzu-labs/object-code-generator-for-k8s)](https://github.com/vmware-tanzu-labs/object-code-generator-for-k8s/blob/main/LICENSE)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/vmware-tanzu-labs/object-code-generator-for-k8s)](https://github.com/vmware-tanzu-labs/object-code-generator-for-k8s/releases)
# Object Code Generator for K8s

Generate source code for unstructured Kubernetes Go types from yaml manifests.

This project is intended for use when scaffolding source code for Go projects
that manage Kubernetes resources.

It can be used in two ways:
1. Imported and used as a package
2. Installed and used as a CLI

## Package

The primary use is as an imported package.  Import the `generate` package and
use it to generate an unstructured Kubernetes object from a yaml manifest.

```go
package main

import (
    "fmt"

    "github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
)

func main() {

    object, err := generate.Generate("path/to/manifest.yaml", "varName")
    if err != nil {
        panic(err)
    }

    fmt.Println(object)
}
```

See `test.go` for a more complete example that uses templating to create a Go
program that will create a Kubernetes deployment resources.

## Command Line Interface

You can also install and use as a CLI.

Install:

```bash
make install
```

Generate object source code from a yaml manifest:

```bash
ocgk generate --manifest-file path/to/manifest.yaml --variable-name varName
```

## Testing

Testing changes to this project involves generating source code for a deployment
resource, then installing that resource in a Kubernetes cluster.  You will need
to have the `KUBECONFIG` env var set that points to a valid kubeconfig for a
running cluster.

Generate source code and run to install deployment in default namespace:

```bash
make test.run
```

Verify the deployment was successfully installed:

```bash
make test.verify
```

Note that this only verifies the deployment was installed.  You may still need
to validate the deployment created includes all intended fields.

Clean up the test deployment:

```bash
make test.clean
```

