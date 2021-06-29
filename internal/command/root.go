// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package command

import (
	"github.com/spf13/cobra"
)

type Root struct {
	Options Options
	Command *cobra.Command
}

func New() *Root {
	rc := &Root{}

	rc.Command = rc.NewCommand()
	rc.AddCommands()

	return rc
}

func (r Root) NewCommand() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands.
	return &cobra.Command{
		Use:   "ocgk",
		Short: "Convert Kubernetes yaml manifests into unstructed Go types",
		Long: `Generate Go source code for unstructured Kubernetes object types from
yaml manifests so that you can manage resources with Go programs.`,
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (r *Root) Execute() {
	cobra.CheckErr(r.Command.Execute())
}

func (r *Root) AddCommands() {
	r.Command.AddCommand(r.GenerateCommand())
}
