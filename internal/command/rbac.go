// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package command

import (
	"github.com/spf13/cobra"

	"github.com/nukleros/gener8s/internal/command/rbac"
)

// GenerateRBACCommand creates the generate subcommand.
func (r *Root) GenerateRBACCommand() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "rbac",
		Short: "Generate various RBAC objects given a set of input manifests.",
		Long: `Pass a set of Kubernetes manifest files for any Kubernetes
object and get RBAC needed to manage it within a cluster (e.g. from a controller).`,
	}

	generateCmd.AddCommand(rbac.MarkersCommand(r.Options))
	generateCmd.AddCommand(rbac.GoCommand(r.Options))
	generateCmd.AddCommand(rbac.YAMLCommand(r.Options))

	return generateCmd
}
