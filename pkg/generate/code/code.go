// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package code

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	ghodss_yaml "github.com/ghodss/yaml"
	"github.com/iancoleman/strcase"
	"github.com/nukleros/gener8s/internal/options"
	"github.com/nukleros/gener8s/pkg/manifests"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ErrTooManyValues = errors.New("only one value struct is allowed")

type element struct {
	Type        string
	Key         string
	Value       string
	IsSeq       bool
	LineComment string
	HeadComment string
	FootComment string
	Elements    elements
}

type object struct {
	VarName  string
	Elements elements
	Source   string
}

type elements []element

func (e *elements) UnmarshalYAML(value *yaml.Node) error {
	e.decodeElements(0, value)

	return nil
}

func (e *elements) decodeElements(factor int, value ...*yaml.Node) {
	for i := 0; i < len(value); i += 1 + factor {
		headComment := strings.Split(value[i].HeadComment, "\n")
		for j := range headComment {
			headComment[j] = strings.Replace(headComment[j], "#", "//", 1)
		}

		footComment := strings.Split(value[i].FootComment, "\n")
		for j := range footComment {
			footComment[j] = strings.Replace(footComment[j], "#", "//", 1)
		}

		hc := strings.Join(headComment, "\n")
		fc := strings.Join(footComment, "\n")

		elem := element{
			Type:        value[i+factor].ShortTag(),
			Key:         value[i].Value,
			LineComment: strings.TrimPrefix(value[i+factor].LineComment, "#"),
			HeadComment: hc,
			FootComment: fc,
		}

		switch value[i+factor].Kind {
		case yaml.DocumentNode:
			e.decodeElements(0, value[i].Content...)
		case yaml.SequenceNode:
			elem.Elements.decodeElements(0, value[i+factor].Content...)

			for i := range elem.Elements {
				elem.Elements[i].IsSeq = true
			}

			*e = append(*e, elem)
		case yaml.MappingNode:
			elem.Elements.decodeElements(1, value[i+factor].Content...)
			*e = append(*e, elem)
		case yaml.ScalarNode:
			elem.Value = value[i+factor].Value
			*e = append(*e, elem)
		case yaml.AliasNode:
			elem.Type = value[i+factor].Alias.ShortTag()
			elem.Value = value[i+factor].Alias.Value
			elem.LineComment = strings.Trim(value[i+factor].Alias.LineComment, "#")
			elem.Elements.decodeElements(1, value[i+factor].Alias.Content...)

			*e = append(*e, elem)
		}
	}
}

// GenerateForManifests generates code for a set of manifest objects.
func GenerateForManifests(manifests *manifests.Manifests, options *options.RBACOptions) (string, error) {
	return "", nil
}

// Generate generates unstructured go types for resources defined in yaml
// manifests.
func Generate(resourceYaml []byte, varName string, values ...interface{}) (string, error) {
	if len(values) > 1 {
		return "", ErrTooManyValues
	} else if len(values) == 1 {
		yamlTemplate, err := template.New("yamlFile").Parse(string(resourceYaml))
		if err != nil {
			return "", fmt.Errorf("unable to parse template in yaml file, %w", err)
		}

		var yamlBuf bytes.Buffer

		if err := yamlTemplate.Execute(&yamlBuf, values[0]); err != nil {
			return "", fmt.Errorf("unable to resolve templating in yaml file, %w", err)
		}

		resourceYaml = yamlBuf.Bytes()
	}

	unstructuredObj := elements{}

	if err := yaml.Unmarshal(resourceYaml, &unstructuredObj); err != nil {
		return "", fmt.Errorf("unable to unmarshal input yaml, %w", err)
	}

	obj := object{
		VarName:  varName,
		Elements: unstructuredObj[0].Elements,
		Source:   string(resourceYaml),
	}

	t, err := template.New("objectTemplate").Funcs(funcMap()).Parse(objTemplate)
	if err != nil {
		return "", fmt.Errorf("unable to parse template, %w", err)
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, obj)
	if err != nil {
		return "", fmt.Errorf("unable to generate go code, %w", err)
	}

	objSource, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("unable to format file, %w", err)
	}

	return string(objSource), nil
}

func escape(str string) string {
	if strings.ContainsAny(str, "\n"+`\`) {
		str = strings.ReplaceAll(str, "`", "` + \"`\" + `")

		if strings.ContainsAny(str, "!!") {
			str = strings.ReplaceAll(str, "!!start", "` +")
			str = strings.ReplaceAll(str, "!!end", "+ `")
		}

		return "`" + str + "`"
	}

	str = strings.ReplaceAll(str, `"`, `\"`)

	if strings.ContainsAny(str, "!!") {
		str = strings.ReplaceAll(str, "!!start", `" +`)
		str = strings.ReplaceAll(str, "!!end", `+ "`)
	}

	return `"` + str + `"`
}

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	f["escape"] = escape

	return f
}

const objTemplate = `
var {{ .VarName }} = &unstructured.Unstructured{
	Object: map[string]interface{}{
		{{- template "element" .Elements }}
	},
}

{{- define "element" }}
	{{- range . }}
		{{- if .HeadComment }}
			{{ .HeadComment }}
		{{- end }}
		{{- if eq .Type "!!null" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": nil,
			{{- else }}
				nil,
			{{- end }}
		{{- else if  or (eq .Type "!!bool") (eq .Type "!!int") }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": {{ .Value -}},  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- else }}
				{{ .Value -}},  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- end }}
		{{- else if eq .Type "!!str" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": {{ escape .Value -}},  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- else }}
				{{ escape .Value -}},  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- end }}
		{{- else if eq .Type "!!var" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": {{ .Value -}},  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- else }}
				{{ .Value -}},  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- end }}
		{{- else if eq .Type "!!map" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": map[string]interface{}{  {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- else }}
				map[string]interface{}{ {{ if .LineComment }}// {{ .LineComment }}{{ end }}
			{{- end }}
				{{- template "element" .Elements }}
			},
		{{- else if eq .Type "!!seq" }}
			"{{ .Key }}": []interface{}{
				{{- template "element" .Elements }}
			},
		{{- end }}
		{{- if .FootComment }}
			{{ .FootComment }}
		{{- end }}
	{{- end }}
{{- end }}
`

// GenerateCode will return the stdout form of unstructured go code, given a set of input
// manifests, in go struct format.
func GenerateCode(files *manifests.Manifests, options *options.RBACOptions, values map[string]interface{}) (string, error) {
	var goString string

	manifestData, err := GenerateYAML(files)
	if err != nil {
		return goString, fmt.Errorf("%w - error generating yaml", err)
	}

	manifest := &manifests.Manifest{Content: []byte(manifestData)}

	extractedManifests := manifest.ExtractManifests()

	for _, resource := range extractedManifests {
		resource = strings.TrimSpace(resource)
		if resource == "" {
			continue
		}

		jsonManifest, err := ghodss_yaml.YAMLToJSON([]byte(resource))
		if err != nil {
			return "", fmt.Errorf("failed to convert YAML to JSON: %v", err)
		}

		// Create an unstructured object from the JSON representation.
		unstructuredObj := &unstructured.Unstructured{}
		if err := json.Unmarshal(jsonManifest, unstructuredObj); err != nil {
			return "", fmt.Errorf("failed to unmarshal JSON into unstructured object: %v", err)
		}

		kind := unstructuredObj.GetKind()
		name := strcase.ToCamel(unstructuredObj.GetName())
		variableName := strcase.ToLowerCamel(kind + name)

		asCode, err := Generate([]byte(resource), variableName, values)
		if err != nil {
			return goString, fmt.Errorf("%w - error generating code for yaml", err)
		}

		if goString == "" {
			goString = fmt.Sprintf("%s\n\n", asCode)
		} else {
			goString = fmt.Sprintf("%s%s\n", goString, asCode)
		}
	}

	return goString, nil
}

// GenerateYAML will return the stdout form of unstructured objects in YAML format given a set of input manifest.
func GenerateYAML(files *manifests.Manifests) (string, error) {

	var resources []string

	for _, manifest := range *files {
		resources = append(resources, manifest.ExtractManifests()...)
	}

	var resourceString string
	for i, resource := range resources {

		if i == 0 {
			resourceString = resource
		} else {
			resourceString = fmt.Sprintf("%s---\n%s", resourceString, resource)
		}
	}

	return resourceString, nil

}
