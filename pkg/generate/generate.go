// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

type element struct {
	Type     string
	Key      string
	Value    string
	IsSeq    bool
	Comment  string
	Elements elements
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
		elem := element{
			Type:    value[i+factor].ShortTag(),
			Key:     value[i].Value,
			Comment: strings.Trim(value[i+factor].LineComment, "#"),
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
			elem.Comment = strings.Trim(value[i+factor].Alias.LineComment, "#")
			elem.Elements.decodeElements(1, value[i+factor].Alias.Content...)

			*e = append(*e, elem)
		}
	}
}

// Generate generates unstructured go types for resources defined in yaml
// manifests.
func Generate(resourceYaml []byte, varName string) (string, error) {
	unstructuredObj := elements{}

	err := yaml.Unmarshal(resourceYaml, &unstructuredObj)
	if err != nil {
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
	if strings.Contains(str, "\n") {
		str = strings.ReplaceAll(str, "`", "` + \"`\" + `")

		return "`" + str + "`,"
	}

	str = strings.ReplaceAll(str, `"`, `\"`)

	return `"` + str + `",`
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
		{{- if eq .Type "!!null" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": nil,
			{{- else }}
				"nil,
			{{- end }}
		{{- else if  or (eq .Type "!!bool") (eq .Type "!!int") }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": {{ .Value -}},  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- else }}
				{{ .Value -}},  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- end }}
		{{- else if eq .Type "!!str" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": {{ escape .Value -}}  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- else }}
				{{ escape .Value -}}  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- end }}
		{{- else if eq .Type "!!map" }}
			{{- if ne .IsSeq true }}
				"{{ .Key }}": map[string]interface{}{  {{ if .Comment }}// {{ .Comment }}{{ end }}
			{{- else }}
				{
			{{- end }}
				{{- template "element" .Elements }}
			},
		{{- else if eq .Type "!!seq" }}
			{{- if not .Elements }}
			    "{{ .Key }}": nil,
			{{- else }}
				{{- if eq (index .Elements 0).Type "!!map" }}
					"{{ .Key }}": []map[string]interface{}{
						{{- template "element" .Elements }}
					},
				{{- else }}
					"{{ .Key }}": []interface{}{
						{{- template "element" .Elements }}
					},
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
{{- end }}
`
