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

// Generate generates unstructured go types for resources defined in yaml
// manifests
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

	//// display json representation of struct for debugging
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
		// unsupported
		fmt.Println("Invalid")
	case reflect.Bool:
		elem = element{
			ValType: "bool",
			Key:     k,
			Value:   strconv.FormatBool(v.(bool)),
		}
	case reflect.Int:
		elem = element{
			ValType: "int",
			Key:     k,
			Value:   strconv.FormatInt(v.(int), 10),
		}
	case reflect.Int8:
		elem = element{
			ValType: "int8",
			Key:     k,
			Value:   strconv.FormatInt(v.(int8), 10),
		}
	case reflect.Int16:
		elem = element{
			ValType: "int16",
			Key:     k,
			Value:   strconv.FormatInt(v.(int16), 10),
		}
	case reflect.Int32:
		elem = element{
			ValType: "int32",
			Key:     k,
			Value:   strconv.FormatInt(v.(int32), 10),
		}
	case reflect.Int64:
		elem = element{
			ValType: "int64",
			Key:     k,
			Value:   strconv.FormatInt(v.(int64), 10),
		}
	case reflect.Uint:
		// unsupported
		fmt.Println("Uint")
	case reflect.Uint8:
		// unsupported
		fmt.Println("Uint8")
	case reflect.Uint16:
		// unsupported
		fmt.Println("Uint16")
	case reflect.Uint32:
		// unsupported
		fmt.Println("Uint32")
	case reflect.Uint64:
		// unsupported
		fmt.Println("Uint64")
	case reflect.Uintptr:
		// unsupported
		fmt.Println("Uintptr")
	case reflect.Float32:
		// unsupported
		fmt.Println("Float32")
	case reflect.Float64:
		// unsupported
		fmt.Println("Float64")
	case reflect.Complex64:
		// unsupported
		fmt.Println("Complex64")
	case reflect.Complex128:
		// unsupported
		fmt.Println("Complex128")
	case reflect.Array:
		// unsupported
		fmt.Println("Array")
	case reflect.Chan:
		// unsupported
		fmt.Println("Chan")
	case reflect.Func:
		// unsupported
		fmt.Println("Func")
	case reflect.Interface:
		// unsupported
		fmt.Println("Interface")
	case reflect.Map:
		elem = element{
			ValType: "map",
			Key:     k,
		}
		for key, value := range v.(map[string]interface{}) {
			newElem := addElement(key, value)
			elem.Elements = append(elem.Elements, *newElem)
		}
	case reflect.Ptr:
		// unsupported
		fmt.Println("Ptr")
	case reflect.Slice:
		sliceVal := v.([]interface{})[0]
		srt := reflect.TypeOf(sliceVal)
		switch srt.Kind() {
		case reflect.String:
			elem = element{
				ValType: "slice-strings",
				Key:     k,
			}
			for _, i := range v.([]interface{}) {
				parentElem := addElement("parent", i)
				parentElem.Parent = true
				elem.Elements = append(elem.Elements, *parentElem)
			}
		default:
			elem = element{
				ValType: "slice-interfaces",
				Key:     k,
			}
			for _, i := range v.([]interface{}) {
				parentElem := addElement("parent", i)
				parentElem.Parent = true
				elem.Elements = append(elem.Elements, *parentElem)
			}
		}
	case reflect.String:
		elem = element{
			ValType: "string",
			Key:     k,
			Value:   v.(string),
		}
	case reflect.Struct:
		// unsupported
		fmt.Println("Struct")
	case reflect.UnsafePointer:
		// unsupported
		fmt.Println("UnsafePointer")
	default:
		// unsupported
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
		{{- else if eq .ValType "bool" }}
			"{{ .Key }}": {{ .Value -}},
		{{- else if eq .ValType "string" }}
			{{- if ne .Parent true }}
				"{{ .Key }}": "{{ .Value -}}",
			{{- else }}
				"{{ .Value -}}",
			{{- end }}
		{{- else if eq .ValType "int" }}
			"{{ .Key }}": int({{ .Value -}}),
		{{- else if eq .ValType "int8" }}
			"{{ .Key }}": int8({{ .Value -}}),
		{{- else if eq .ValType "int16" }}
			"{{ .Key }}": int16({{ .Value -}}),
		{{- else if eq .ValType "int32" }}
			"{{ .Key }}": int32({{ .Value -}}),
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
		{{- else if eq .ValType "slice-strings" }}
			"{{ .Key }}": []string{
				{{- template "element" .Elements }}
			},
		{{- else if eq .ValType "slice-interfaces" }}
			"{{ .Key }}": []map[string]interface{}{
				{{- template "element" .Elements }}
			},
		{{- end }}
	{{- end }}
{{- end }}
`
