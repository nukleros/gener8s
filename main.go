package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func main() {

	manifestFile, _ := filepath.Abs("sample-ns.yaml")
	yamlFile, err := ioutil.ReadFile(manifestFile)
	if err != nil {
		panic(err)
	}

	var ns v1.Namespace

	err = yaml.Unmarshal(yamlFile, &ns)
	if err != nil {
		panic(err)
	}

	t, err := template.New("thing").Parse(nsTemplate)
	if err != nil {
		panic(err)
	}
	err = t.Execute(os.Stdout, ns)
	if err != nil {
		panic(err)
	}

}

const nsTemplate = `
var ns = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: {{ .ObjectMeta.Name }},
	},
}
`
