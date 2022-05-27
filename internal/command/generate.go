// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/nukleros/gener8s/pkg/generate"
)

type Options struct {
	manifestFilepath string
	variableName     string
	valuesFilePath   string
}

// GenerateGoCommand creates the generate subcommand.
func (r *Root) GenerateGoCommand() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "go",
		Short: "Generate Go source code for Kubernetes object from yaml",
		Long: `Pass a manifest file that contains valid yaml for any Kubernetes
object and get source code for an unstructured Kubernetes object type.`,
		Example: `
# generate unstructured go code for a kubernetes object
gener8s go -m /path/to/rbac.yaml
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestFile, err := filepath.Abs(r.Options.manifestFilepath)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			yamlContent, err := ioutil.ReadFile(manifestFile)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			var values map[string]interface{}

			if r.Options.valuesFilePath != "" {
				valuesFile, vErr := ioutil.ReadFile(r.Options.valuesFilePath)
				if err != nil {
					return fmt.Errorf("%w", vErr)
				}

				if vErr := yaml.Unmarshal(valuesFile, &values); err != nil {
					return fmt.Errorf("%w", vErr)
				}
			}

			source, err := generate.Generate(yamlContent, r.Options.variableName, values)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			os.Stdout.WriteString(source)

			return nil
		},
	}

	generateCmd.Flags().StringVarP(
		&r.Options.manifestFilepath,
		"manifest-file",
		"m",
		"",
		"path to manifest file containing resource definition",
	)

	generateCmd.Flags().StringVarP(
		&r.Options.variableName,
		"variable-name",
		"v",
		"object",
		"variable name for resource object",
	)

	generateCmd.Flags().StringVarP(
		&r.Options.valuesFilePath,
		"values-file",
		"f",
		"",
		"yaml file with values to insert into fields with !!tpl tags",
	)

	cobra.CheckErr(generateCmd.MarkFlagRequired("manifest-file"))

	return generateCmd
}
