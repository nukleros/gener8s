package rbac

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nukleros/gener8s/internal/options"
	"github.com/nukleros/gener8s/pkg/generate/rbac"
	"github.com/nukleros/gener8s/pkg/manifests"
)

var ErrUnsupportedGenerateOption = errors.New("unsupported generate option")

// addFlags adds the common rbac flags.
func addFlags(cmd *cobra.Command, options *options.RBACOptions) {
	cmd.Flags().StringArrayVarP(
		&options.ManifestFilepaths,
		"manifest-files",
		"m",
		[]string{},
		"path to manifest files containing resource definition; may include globbing",
	)

	cmd.Flags().StringArrayVar(
		&options.Verbs,
		"verbs",
		rbac.DefaultResourceVerbs(),
		"verbs needed for the rbac generation (applies to all objects passed in with the -m flag)",
	)

	cobra.CheckErr(cmd.MarkFlagRequired("manifest-files"))
}

// run adds the run function.
func run(cliOptions *options.RBACOptions, rbacOption options.GenerateOption) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		manifests, err := manifests.ExpandManifests("", cliOptions.ManifestFilepaths)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// load manifest content for each manifest
		for _, manifest := range *manifests {
			if err = manifest.LoadContent(); err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		var stdout string

		switch rbacOption {
		case options.WithYAML:
			stdout, err = rbac.GenerateYAML(manifests, cliOptions)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		case options.WithGo:
			stdout, err = rbac.GenerateCode(manifests, cliOptions)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		case options.WithMarkers:
			stdout, err = rbac.GenerateMarkers(manifests, cliOptions)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		default:
			return ErrUnsupportedGenerateOption
		}

		os.Stdout.WriteString(stdout)

		return nil
	}
}
