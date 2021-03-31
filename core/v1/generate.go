package v1

import (
	"os"
	"text/template"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const nsTemplate = `
var ns = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: {{ .ObjectMeta.Name }},
	},
}
`

// genNamespace generates go type source for namespace
func GenNamespace(yamlFile []byte) error {

	var ns v1.Namespace

	err := yaml.Unmarshal(yamlFile, &ns)
	if err != nil {
		return err
	}

	t, err := template.New("thing").Parse(nsTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(os.Stdout, ns)
	if err != nil {
		return err
	}

	return nil
}
