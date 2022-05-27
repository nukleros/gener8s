package rbac

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nukleros/gener8s/internal/options"
	"github.com/nukleros/gener8s/pkg/generate/rbac"
	"github.com/nukleros/gener8s/pkg/manifests"
)

// MarkersCommand creates the `rbac markers` subcommand.
func MarkersCommand(options *options.Options) *cobra.Command {
	markersCmd := &cobra.Command{
		Use:   "kubebuilder-markers",
		Short: "Generate RBAC kubebuilder markers needed to manage Kubernetes manifests.",
		Long: `Pass a set of Kubernetes manifest files for any Kubernetes
object and get kubebuilder RBAC markers needed for controller-gen.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifests, err := manifests.ExpandManifests("", options.ManifestFilepaths)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			// load manifest content for each manifest
			for _, manifest := range *manifests {
				if err = manifest.LoadContent(); err != nil {
					return fmt.Errorf("%w", err)
				}
			}

			// generate the kubebuilder markers
			source, err := rbac.Generate(manifests, rbac.WithMarkers)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			os.Stdout.WriteString(source)

			return nil
		},
	}

	markersCmd.Flags().StringArrayVarP(
		&options.ManifestFilepaths,
		"manifest-files",
		"m",
		[]string{},
		"path to manifest files containing resource definition; may include globbing",
	)

	cobra.CheckErr(markersCmd.MarkFlagRequired("manifest-files"))

	return markersCmd
}
