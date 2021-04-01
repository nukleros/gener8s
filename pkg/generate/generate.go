package generate

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type element struct {
	ValType  string
	Key      string
	Value    string
	Parent   bool
	Elements []element
}

type object struct {
	VarName  string
	Elements []element
}

func Generate(filename, varName string) (string, error) {

	manifestFile, _ := filepath.Abs(filename)
	yamlFile, err := ioutil.ReadFile(manifestFile)
	if err != nil {
		return "", err
	}

	unstructuredObj := unstructured.Unstructured{}

	err = yaml.Unmarshal(yamlFile, &unstructuredObj)
	if err != nil {
		return "", err
	}

	obj := object{VarName: varName}

	for k, v := range unstructuredObj.Object {
		elem := addElement(k, v)
		obj.Elements = append(obj.Elements, *elem)
	}

	//// display json repr of struct for debugging
	//objJson, err := json.MarshalIndent(obj, ``, `  `)
	//if err != nil {
	//	return "", err
	//}
	//fmt.Println(string(objJson))

	t, err := template.New("objectTemplate").Parse(objTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, obj); err != nil {
		return "", err
	}

	objSource, err := format.Source(buf.Bytes())
	if err != nil {
		return "", err
	}

	return string(objSource), nil
}

func addElement(k string, v interface{}) *element {

	var elem element

	if v == nil {
		elem = element{
			ValType: "nil",
			Key:     k,
			Value:   "nil",
		}
		return &elem
	}
	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.Invalid:
		fmt.Println("Invalid")
	case reflect.Bool:
		fmt.Println("Bool")
	case reflect.String:
		elem = element{
			ValType: "string",
			Key:     k,
			Value:   v.(string),
		}
	case reflect.Int:
		fmt.Println("Int")
	case reflect.Int64:
		elem = element{
			ValType: "int64",
			Key:     k,
			Value:   strconv.FormatInt(v.(int64), 10),
		}
	case reflect.Map:
		elem = element{
			ValType: "map",
			Key:     k,
		}
		for key, value := range v.(map[string]interface{}) {
			newElem := addElement(key, value)
			elem.Elements = append(elem.Elements, *newElem)
		}
	case reflect.Slice:
		elem = element{
			ValType: "slice",
			Key:     k,
		}
		for _, i := range v.([]interface{}) {
			parentElem := addElement("parent", i)
			parentElem.Parent = true
			elem.Elements = append(elem.Elements, *parentElem)
		}
	case reflect.Array:
		fmt.Println("Array")
	default:
		fmt.Println("default")
	}

	return &elem
}

const objTemplate = `
var {{ .VarName }} = &unstructured.Unstructured{
	Object: map[string]interface{}{
		{{- template "element" .Elements }}
	},
}

{{- define "element" }}
	{{- range . }}
		{{- if eq .ValType "nil" }}
			"{{ .Key }}": nil,
		{{- else if eq .ValType "string" }}
			"{{ .Key }}": "{{ .Value -}}",
		{{- else if eq .ValType "int64" }}
			"{{ .Key }}": int64({{ .Value -}}),
		{{- else if eq .ValType "map" }}
			{{- if ne .Parent true }}
				"{{ .Key }}": map[string]interface{}{
			{{- else }}
				{
			{{- end }}
				{{- template "element" .Elements }}
			},
		{{- else if eq .ValType "slice" }}
			"{{ .Key }}": []map[string]interface{}{
				{{- template "element" .Elements }}
			},
		{{- end }}
	{{- end }}
{{- end }}
`
