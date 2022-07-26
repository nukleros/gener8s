package rbac

import (
	"github.com/spf13/cobra"

	"github.com/nukleros/gener8s/internal/options"
)

// MarkersCommand creates the `rbac markers` subcommand.
func MarkersCommand(cliOptions *options.RBACOptions) *cobra.Command {
	markersCmd := &cobra.Command{
		Use:   "markers",
		Short: "Generate RBAC kubebuilder markers.",
		Long: `Pass a set of Kubernetes manifest files for any Kubernetes
object and get RBAC resources as kubebuilder markers (needed for controller-gen).`,
		RunE: run(cliOptions, options.WithMarkers),
	}

	// add flags
	addFlags(markersCmd, cliOptions)

	return markersCmd
}
