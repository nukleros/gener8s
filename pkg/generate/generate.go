package generate

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"go/format"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type element struct {
	ValType  string
	Key      string
	Value    string
	Parent   bool
	Comment  string
	Elements []element
}

type object struct {
	VarName  string
	Elements []element
}

// Generate generates unstructured go types for resources defined in yaml
// manifests
func Generate(resourceYaml []byte, varName string) (string, error) {

	commentedYaml, err := captureComments(string(resourceYaml))
	if err != nil {
		return "", err
	}

	unstructuredObj := unstructured.Unstructured{}

	err = yaml.Unmarshal([]byte(commentedYaml), &unstructuredObj)
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

	var comment string
	if strings.Contains(k, "+comment") {
		k, comment = goComment(k)
	}

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
			Comment: expandColonSpace(comment),
		}
	case reflect.Int:
		elem = element{
			ValType: "int",
			Key:     k,
			Value:   strconv.FormatInt(int64(v.(int)), 10),
			Comment: expandColonSpace(comment),
		}
	case reflect.Int8:
		elem = element{
			ValType: "int8",
			Key:     k,
			Value:   strconv.FormatInt(int64(v.(int8)), 10),
			Comment: expandColonSpace(comment),
		}
	case reflect.Int16:
		elem = element{
			ValType: "int16",
			Key:     k,
			Value:   strconv.FormatInt(int64(v.(int16)), 10),
			Comment: expandColonSpace(comment),
		}
	case reflect.Int32:
		elem = element{
			ValType: "int32",
			Key:     k,
			Value:   strconv.FormatInt(int64(v.(int32)), 10),
			Comment: expandColonSpace(comment),
		}
	case reflect.Int64:
		elem = element{
			ValType: "int64",
			Key:     k,
			Value:   strconv.FormatInt(v.(int64), 10),
			Comment: expandColonSpace(comment),
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
			Comment: expandColonSpace(comment),
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
				Comment: expandColonSpace(comment),
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
				Comment: expandColonSpace(comment),
			}
			for _, i := range v.([]interface{}) {
				parentElem := addElement("parent", i)
				parentElem.Parent = true
				elem.Elements = append(elem.Elements, *parentElem)
			}
		}
	case reflect.String:
		val := v.(string)
		if strings.Contains(val, "+comment") {
			val, comment = goComment(val)
		}
		elem = element{
			ValType: "string",
			Key:     k,
			Value:   strings.Trim(val, " "),
			Comment: expandColonSpace(comment),
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

// captureComments captures the comment and adds it to the key so that it is not
// lost during yaml unmarshalling
func captureComments(rawContent string) (string, error) {

	lines := strings.Split(string(rawContent), "\n")
	for i, line := range lines {
		if containsComment(line) {
			commentedLine, err := processComments(line)
			if err != nil {
				return "", err
			}
			lines[i] = commentedLine
		}
	}

	return strings.Join(lines, "\n"), nil
}

// containsComment returns true if there's a comment that is not a part of any
// key or value
// TODO: be more intelligent about "#" marks inside keys or values that aren't
// actually comments
func containsComment(line string) bool {

	if len(line) > 0 && strings.TrimLeft(line, " ")[:1] == "#" {
		// first char is "#" a standalone comment - these get skipped
		return false
	} else {
		return strings.Contains(line, "#")
	}
}

// processComments splits a single line into its keys and values as needed, then
// sends the line contents to get the comment extracted
func processComments(line string) (string, error) {

	// for array value lines e.g. `- value # comment:with:colons`
	// the line should not be split on the colon as it does not effectively
	// split the line into keys and values
	keyVal := true
	colonEncountered := false
	for _, char := range line {
		if char == ':' {
			colonEncountered = true
		} else if char == '#' && colonEncountered == false {
			keyVal = false
			break
		}
	}

	var commentEncodedLine string
	var keyValueArray []string

	// only split on the colon for key value lines e.g. `foo: bar # comment`
	if keyVal {
		keyValueArray = strings.Split(line, ":")
	} else {
		keyValueArray = []string{line}
	}

	if len(keyValueArray) > 2 {
		keyValueArray = processValueColons(keyValueArray)
		commentEncodedLine = extractComment(keyValueArray)
	} else {
		commentEncodedLine = extractComment(keyValueArray)
	}
	return commentEncodedLine, nil
}

// processValueColons accounts for colons in values once line is split to
// extract keys and values
func processValueColons(lineArray []string) []string {

	var value string
	for i, v := range lineArray {
		if i == 1 {
			value = v
		} else if i > 1 {
			value = value + ":" + v
		}
	}

	return []string{lineArray[0], value}
}

// extractComment splits around the comment symbol and adds the comment to the
// key
func extractComment(lineArray []string) string {

	var lineContent string
	if len(lineArray) > 1 {
		// key value pair OR key without value
		key := lineArray[0]
		value := lineArray[1]
		commentArray := strings.Split(value, "#")
		comment := commentArray[1]
		comment = collapseColonSpace(comment)
		comment = fmt.Sprintf("+comment(%s)", comment)
		key = key + comment
		lineContent = fmt.Sprintf("%s: %s", key, value)
	} else {
		// no colon, e.g. array value
		lineContent = lineArray[0]
		lineContentArray := strings.Split(lineContent, "#")
		comment := lineContentArray[1]
		comment = collapseColonSpace(comment)
		comment = fmt.Sprintf("+comment(%s)", comment)
		lineContent = fmt.Sprintf("%s%s", lineContentArray[0], comment)
	}

	return lineContent
}

// collapseColonSpace replaces ": " with ":~" so that yaml unmarshalling won't
// interpret this as a yaml key-value pair
func collapseColonSpace(input string) string {
	return strings.Replace(input, ": ", ":~", -1)
}

// expandColonSpace does the opposite of collapseColonSpace - it replaces ":~"
// with ": " to restore the original string post-yaml unmarshalling
func expandColonSpace(input string) string {
	return strings.Replace(input, ":~", ": ", -1)
}

// goComment pulls out the comments for Go source
func goComment(key string) (string, string) {

	splitKey := strings.Split(key, "(")
	comment := strings.TrimSuffix(splitKey[1], ")")
	k := strings.TrimSuffix(splitKey[0], "+comment")

	return k, comment
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
			"{{ .Key }}": {{ .Value -}},  {{ if .Comment }}// {{ .Comment }}{{ end }}
		{{- else if eq .ValType "string" }}
			{{- if ne .Parent true }}
				"{{ .Key }}": "{{ .Value -}}",  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- else }}
				"{{ .Value -}}",  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- end }}
		{{- else if eq .ValType "int" }}
			"{{ .Key }}": int({{ .Value -}}),  {{ if .Comment }}// {{ .Comment }}{{ end }}
		{{- else if eq .ValType "int8" }}
			"{{ .Key }}": int8({{ .Value -}}),  {{ if .Comment }}// {{ .Comment }}{{ end }}
		{{- else if eq .ValType "int16" }}
			"{{ .Key }}": int16({{ .Value -}}),  {{ if .Comment }}// {{ .Comment }}{{ end }}
		{{- else if eq .ValType "int32" }}
			"{{ .Key }}": int32({{ .Value -}}),  {{ if .Comment }}// {{ .Comment }}{{ end }}
		{{- else if eq .ValType "int64" }}
			"{{ .Key }}": int64({{ .Value -}}),  {{ if .Comment }}// {{ .Comment }}{{ end }}
		{{- else if eq .ValType "map" }}
			{{- if ne .Parent true }}
				"{{ .Key }}": map[string]interface{}{  {{ if .Comment }}// {{ .Comment }}{{ end }}
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
