// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package main

import "github.com/nukleros/gener8s/internal/command"

func main() {
	gener8s := command.New()

	gener8s.Execute()
}
