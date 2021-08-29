// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package generate

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
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

// Generate generates unstructured go types for resources defined in yaml
// manifests.
func Generate(resourceYaml []byte, varName string, values ...interface{}) (string, error) {
	if len(values) > 1 {
		return "", ErrTooManyValues
	}

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

	t, err := template.New("objectTemplate").Funcs(funcMap(values[0])).Parse(objTemplate)
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

		return "`" + str + "`"
	}

	str = strings.ReplaceAll(str, `"`, `\"`)

	return `"` + str + `"`
}

func handleNestedTemplates(values interface{}) func(str string) (string, error) {
	return func(str string) (string, error) {
		var buf bytes.Buffer

		t, err := template.New(str).Parse(str)
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}

		if err := t.Execute(&buf, values); err != nil {
			return "", fmt.Errorf("%w", err)
		}

		return buf.String(), nil
	}
}

func funcMap(value interface{}) template.FuncMap {
	f := sprig.TxtFuncMap()
	f["escape"] = escape
	f["interpolate"] = handleNestedTemplates(value)

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
			{{- if eq .Type "!!tpl" }}
				{{ interpolate .HeadComment }}
			{{- else }}
				{{ .HeadComment }}
			{{- end }}
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
		{{- else if eq .Type "!!tpl" }}
			{{- if ne .IsSeq true }}
				"{{ interpolate .Key }}": {{ interpolate .Value -}},  {{ if .LineComment }}// {{ interpolate .LineComment }}{{ end }}
			{{- else }}
				{{ interpolate .Value -}},  {{ if .LineComment }}// {{ interpolate .LineComment }}{{ end }}
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
			{{- if eq .Type "!!tpl" }}
				{{ interpolate .FootComment }}
			{{- else }}
				{{ .FootComment }}
			{{- end }}
		{{- end }}
	{{- end }}
{{- end }}
`
