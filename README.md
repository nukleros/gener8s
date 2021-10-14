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
    "io/ioutil"

    "github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
)

func main() {

    manifestYaml, err := ioutil.ReadFile("path/to/yaml/file")
    if err != nil {
        panic(err)
    }

    object, err := generate.Generate(manifestYaml, "varName")
    if err != nil {
        panic(err)
    }

    fmt.Println(object)
}
```

See `cmd/ocgk/main_test.go` for a more complete example that uses templating to create a Go
program that will create a Kubernetes deployment resource in a cluster.

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


## Templating

You can also resolve templating within the manifests, values may be given via
the optional values parameter or with the -f flag when using the CLI. This can
be useful when dealing with multiple layers of code generatation, or for
generating code with variable references.

## Variable References

Sometimes you may want to generate code with variable references. To tell the
generator a value is a variable, you may use a special `!!var` yaml tag on that value.


## Variable Reference Inside a string
Sometimes to may want to generate code with a variable reference inside a string. To tell the 
generator a value contains a variable inside it. Inside the value you may the special tags `!!start` to mark the start of the variable and `!!end` to mark the end. the generator will automatically interpolate these and escape quotation marks appropriately
## Example

in this example we will combine templating, variables, and nested variables.  Note that all
these features are optional.  They can be used independently,
together, or not at all.

Example manifest:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: '{{ .Name }}'  # template variable
spec:
    replicas: 2
    selector:
        matchLabels:
            app: !!var webstoreLabel  # variable reference
    template:
        metadata:
            labels:
                app: '{{ .Label }}'  # templated reference
        spec:
            containers:
              - name: webstore-container
                image: my.private.repo/!!start image !!end  # nested variable reference
                ports:
                  - containerPort: 8080
```

Example values file:

```yaml
Name: MyName
Label: webstoreLabel
Image: variable.With.Image.Value
```

This manifest and values file will produce:

```go
var test = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "MyName",
		},
		"spec": map[string]interface{}{
			"replicas": 2,
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": webstoreLabel,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": webstoreLabel,
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "webstore-container",
							"image": "my.private.repo/" + variable.With.Image.Value + "",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 8080,
								},
							},
						},
					},
				},
			},
		},
	},
}
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

