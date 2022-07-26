package rbac

import (
	"github.com/spf13/cobra"

	"github.com/nukleros/gener8s/internal/options"
)

// YAMLCommand creates the `rbac markers` subcommand.
func YAMLCommand(cliOptions *options.RBACOptions) *cobra.Command {
	yamlCmd := &cobra.Command{
		Use:   "yaml",
		Short: "Generate RBAC as YAML manifests.",
		Long: `Pass a set of Kubernetes manifest files for any Kubernetes
object and get RBAC resources as YAML manifests.`,
		RunE: run(cliOptions, options.WithYAML),
	}

	// add flags
	addFlags(yamlCmd, cliOptions)

	// add flags specific to the yaml subcommand
	yamlCmd.Flags().StringVar(
		&cliOptions.RoleName,
		"role-name",
		"manager-role",
		"name of the role(s) to use for the generated yaml objects",
	)

	yamlCmd.Flags().BoolVar(
		&cliOptions.UseResourceNames,
		"use-resource-names",
		false,
		"lock down rbac generation to use the 'resourceNames' field for generated rbac markers",
	)

	return yamlCmd
}
