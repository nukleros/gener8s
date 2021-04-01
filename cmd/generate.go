/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gitlab.eng.vmware.com/landerr/k8s-object-code-generator/pkg/generate"
)

var (
	manifestFile string
	variableName string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go source code for Kubernetes object from yaml",
	Long: `Pass a manifest file that contains valid yaml for any Kubernetes
object and get source code for an unstructured Kubernetes object type.`,
	Run: func(cmd *cobra.Command, args []string) {
		source, err := generate.Generate(manifestFile, variableName)
		if err != nil {
			panic(err)
		}
		fmt.Println(source)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&manifestFile, "manifest-file", "m", "", "manifest file containing resource definition")
	generateCmd.Flags().StringVarP(&variableName, "variable-name", "v", "object", "variable name for resource object")
	generateCmd.MarkFlagRequired("manifest-file")
}
