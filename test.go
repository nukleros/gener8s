package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
)

var manifestFile string

type source struct {
	Object string
}

func main() {

	var manifestPath string
	var outputPath string

	flag.StringVar(&manifestPath, "manifest", "sample/deploy-part.yaml", "path to resource manifest")
	flag.StringVar(&outputPath, "output", "/tmp/kocg-test.go", "path to output go source code")

	flag.Parse()

	t, err := template.New("testTemplate").Parse(testTemplate)
	if err != nil {
		panic(err)
	}

	manifestYaml, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}

	object, err := generate.Generate(manifestYaml, "deployment")
	if err != nil {
		panic(err)
	}

	src := source{Object: object}
	var buf bytes.Buffer
	if err = t.Execute(&buf, src); err != nil {
		panic(err)
	}
	fileSource, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Write(fileSource)
}

const testTemplate = `
package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	namespace := "default"

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	{{ .Object }}

	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetName())
}
`
