// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/nukleros/gener8s/pkg/generate/code"
	"github.com/nukleros/gener8s/pkg/manifests"
)

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

			var values map[string]interface{}

			if r.Options.ValuesFilePath != "" {
				valuesFile, vErr := os.ReadFile(r.Options.ValuesFilePath)
				if vErr != nil {
					return fmt.Errorf("%w", vErr)
				}

				if vErr := yaml.Unmarshal(valuesFile, &values); vErr != nil {
					return fmt.Errorf("%w", vErr)
				}
			}

			manifests, err := manifests.ExpandManifests("", r.Options.ManifestFilepaths)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			// load manifest content for each manifest
			for _, manifest := range *manifests {
				if err = manifest.LoadContent(); err != nil {
					return fmt.Errorf("%w", err)
				}
			}

			source, err := code.GenerateCode(manifests, r.Options)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			os.Stdout.WriteString(source)

			return nil
		},
	}

	generateCmd.Flags().StringArrayVarP(
		&r.Options.ManifestFilepaths,
		"manifest-files",
		"m",
		[]string{},
		"path to manifest files containing resource definition; may include globbing",
	)

	generateCmd.Flags().StringVarP(
		&r.Options.VariableName,
		"variable-name",
		"v",
		"object",
		"variable name for resource object",
	)

	generateCmd.Flags().StringVarP(
		&r.Options.ValuesFilePath,
		"values-file",
		"f",
		"",
		"yaml file with values to insert into fields with !!tpl tags",
	)

	cobra.CheckErr(generateCmd.MarkFlagRequired("manifest-files"))

	return generateCmd
}
