// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package main

import "github.com/vmware-tanzu-labs/object-code-generator-for-k8s/internal/command"

func main() {
	ocgk := command.New()

	ocgk.Execute()
}
