// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package command

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
)

type Options struct {
	manifestFilepath string
	variableName     string
}

// GenerateCommand creates the generate subcommand.
func (r *Root) GenerateCommand() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Go source code for Kubernetes object from yaml",
		Long: `Pass a manifest file that contains valid yaml for any Kubernetes
object and get source code for an unstructured Kubernetes object type.`,
		Run: func(cmd *cobra.Command, args []string) {
			manifestFile, err := filepath.Abs(r.Options.manifestFilepath)
			if err != nil {
				panic(err)
			}

			yamlContent, err := ioutil.ReadFile(manifestFile)
			if err != nil {
				panic(err)
			}

			source, err := generate.Generate(yamlContent, r.Options.variableName)
			if err != nil {
				panic(err)
			}

			os.Stdout.WriteString(source)
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

	cobra.CheckErr(generateCmd.MarkFlagRequired("manifest-file"))

	return generateCmd
}
