// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/nukleros/gener8s/pkg/manifests"
)

type GenerateOptions int

const (
	WithMarkers GenerateOptions = iota
	WithOnlyYAML
	WithOnly
)

// rbacWorkloadProcessor is an interface which implements processing of rbac rules
// for individual workloads (e.g. standalone, collection, component).
type rbacWorkloadProcessor interface {
	IsComponent() bool

	GetDomain() string
	GetAPIGroup() string
	GetAPIVersion() string
	GetAPIKind() string
}

// rbacRuleProcessor is an interface which implements processing of individual
// rbac rules.
type rbacRuleProcessor interface {
	addTo(*Rules)
}

const (
	coreGroup         = "core"
	kubebuilderPrefix = "// +kubebuilder:rbac"
)

// defaultResourceVerbs is a helper function to define the default verbs that are allowed
// for resources that are managed by the scaffolded controller.
func defaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
}

// defaultStatusVerbs is a helper function to define the default verbs which get placed
// onto resources that have a /status suffix.
func defaultStatusVerbs() []string {
	return []string{
		"get", "update", "patch",
	}
}

// knownIrregulars is a helper function to define known irregular kinds and their
// expected formats.
//   - keys   = found values
//   - values = proper values
func knownIrregulars() map[string]string {
	return map[string]string{
		"resourcequota": "resourcequotas",
	}
}

// Generate will return the stdout form of rbac objects.
func Generate(files *manifests.Manifests, options ...GenerateOptions) (string, error) {
	var rbacString string

	for _, manifest := range *files {
		// decode manifest into unstructured data type
		var manifestObject unstructured.Unstructured

		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		if err := runtime.DecodeInto(decoder, manifest.Content, &manifestObject); err != nil {
			return "", fmt.Errorf("%w; unable to decode object in manifest file %s", err, manifest.Filename)
		}

		// determine the rbac rules for this resource
		rules, err := ForResource(&manifestObject)
		if err != nil {
			return "", err
		}

		// loop through the rules and append to the string
		for _, rule := range *rules {
			rbacString = fmt.Sprintf("%s\n%s", rbacString, rule.ToMarker())
		}
	}

	return rbacString, nil
}

// ForResource will return a set of rules for a particular kubernetes resource.  This includes
// a rule for the resource itself, in addition to adding particular rules for whatever
// roles and cluster roles are requesting.  This is because the controller needs to have
// permissions to manage the children that roles and cluster roles are requesting.
func ForResource(manifest *unstructured.Unstructured) (*Rules, error) {
	rules := &Rules{}

	if err := rules.addForResource(manifest); err != nil {
		return rules, err
	}

	return rules, nil
}

// ForResources will return a set of rules for particular kubernetes resources.  See ForResource
// for more information as this is the same methodology used.
func ForResources(manifests ...*unstructured.Unstructured) (*Rules, error) {
	rules := &Rules{}

	for _, manifest := range manifests {
		if err := rules.addForResource(manifest); err != nil {
			return rules, err
		}
	}

	return rules, nil
}

// ForWorkloads will return a set of rules for a particular set of workloads.  It should be noted that
// this only returns the specific rules for the actual workload and not the managed resources.  See
// ForManifest for details on the rules for a particular manifest.
func ForWorkloads(workloads ...rbacWorkloadProcessor) *Rules {
	rules := &Rules{}

	// for each of the workloads passed in, add a rule to the set of rules
	for i := range workloads {
		rules.addForWorkload(workloads[i])
	}

	return rules
}

// getGroup returns the group in the proper format as expected by rbac markers.
func getGroup(group string) string {
	if group == "" {
		return coreGroup
	}

	return group
}

// getFieldString returns an array of fields in string format.
func getFieldString(fields []string) string {
	return strings.Join(fields, ";")
}

// getResource gets the resource properly formatted for an rbac rule given the kind
// of resource.  For regular rules, the kind comes in as expected, but for role
// rules, this could come in as an asterisk so it has to be specially handled.
func getResource(kind string) string {
	rbacResource := strings.Split(kind, "/")

	if rbacResource[0] == "*" {
		kind = "*"
	} else {
		kind = getPlural(rbacResource[0])
	}

	if len(rbacResource) > 1 {
		kind = fmt.Sprintf("%s/%s", kind, rbacResource[1])
	}

	return kind
}

// getPlural will transform known irregulars into a proper type for rbac
// rules.
func getPlural(kind string) string {
	plural := resource.RegularPlural(kind)

	if knownIrregulars()[plural] != "" {
		return knownIrregulars()[plural]
	}

	return plural
}

// valueFromInterface attempts to retrieve a value from an interface as a map.
func valueFromInterface(in interface{}, key string) (out interface{}) {
	switch asType := in.(type) {
	case map[interface{}]interface{}:
		out = asType[key]
	case map[string]interface{}:
		out = asType[key]
	case map[interface{}][]interface{}:
		out = asType[key]
	case map[string][]interface{}:
		out = asType[key]
	}

	return out
}
