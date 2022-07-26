package rbac

import (
	"github.com/spf13/cobra"

	"github.com/nukleros/gener8s/internal/options"
)

// GoCommand creates the `rbac markers` subcommand.
func GoCommand(cliOptions *options.RBACOptions) *cobra.Command {
	goCmd := &cobra.Command{
		Use:   "go",
		Short: "Generate RBAC as unstructured Go.",
		Long: `Pass a set of Kubernetes manifest files for any Kubernetes
object and get RBAC resources as unstructred Go YAML.`,
		RunE: run(cliOptions, options.WithGo),
	}

	// add flags
	addFlags(goCmd, cliOptions)

	// add flags specific to the go subcommand
	goCmd.Flags().StringVar(
		&cliOptions.VariableName,
		"variable-name",
		"resourceObj",
		"name of the variable that is used for the object when generating the go code",
	)

	goCmd.Flags().StringVar(
		&cliOptions.RoleName,
		"role-name",
		"manager-role",
		"name of the role(s) to use for the generated go struct objects",
	)

	goCmd.Flags().BoolVar(
		&cliOptions.UseResourceNames,
		"use-resource-names",
		false,
		"lock down rbac generation to use the 'resourceNames' field for generated rbac markers",
	)

	return goCmd
}
