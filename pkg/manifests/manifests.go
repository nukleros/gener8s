// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package manifests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nukleros/gener8s/pkg/utils"
)

var ErrProcessManifest = errors.New("error processing manifest file")

type ManifestOptions int

const (
	WithParentPath ManifestOptions = iota
)

// Manifest represents a single input manifest for a given config.
type Manifest struct {
	Content          []byte
	Filename         string
	RelativeFilename string
}

// Manifests represents a collection of manifests.
type Manifests []*Manifest

// ExpandManifests expands manifests from its globbed pattern and return the resultant manifest
// filenames from the glob.
func ExpandManifests(parentPath string, manifestPaths []string) (*Manifests, error) {
	var manifests Manifests

	for i := range manifestPaths {
		files, err := utils.Glob(filepath.Join(parentPath, manifestPaths[i]))
		if err != nil {
			return &Manifests{}, fmt.Errorf("failed to process glob pattern matching, %w", err)
		}

		for f := range files {
			var rf string
			if parentPath != "" {
				rf, err = filepath.Rel(parentPath, files[f])
				if err != nil {
					return &Manifests{}, fmt.Errorf("unable to determine relative file path, %w", err)
				}
			} else {
				rf = manifestPaths[i]
			}

			manifest := &Manifest{Filename: files[f], RelativeFilename: rf}
			manifests = append(manifests, manifest)
		}
	}

	return &manifests, nil
}

// ExtractManifests extracts the manifests as YAML strings from a manifest with
// existing manifest content.
func (manifest *Manifest) ExtractManifests() []string {
	var manifests []string

	manifestYaml := strings.Split(string(manifest.Content), "---")

	for _, object := range manifestYaml {
		object = strings.TrimSpace(object)
		if object == "" {
			continue
		}
		manifests = append(manifests, object)
	}

	return manifests
}

// LoadContent sets the Content field of the manifest in raw format as []byte.
func (manifest *Manifest) LoadContent() error {
	manifestContent, err := os.ReadFile(manifest.Filename)
	if err != nil {
		return fmt.Errorf("%w; %s for manifest file %s", err, ErrProcessManifest, manifest.Filename)
	}

	manifest.Content = manifestContent

	return nil
}
