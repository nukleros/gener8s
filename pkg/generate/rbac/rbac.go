// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-tools/pkg/rbac"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/nukleros/gener8s/internal/options"
	"github.com/nukleros/gener8s/pkg/generate/code"
	"github.com/nukleros/gener8s/pkg/manifests"
)

const (
	coreGroup         = "core"
	kubebuilderPrefix = "// +kubebuilder:rbac"
)

// rbacWorkloadProcessor is an interface which implements processing of rbac rules
// for individual workloads (e.g. standalone, collection, component).
type rbacWorkloadProcessor interface {
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

// ruleKey represents the resources and non-resources a Rule applies.
type ruleKey struct {
	Groups        string
	Resources     string
	ResourceNames string
	URLs          string
}

func (key ruleKey) String() string {
	return fmt.Sprintf("%s + %s + %s + %s", key.Groups, key.Resources, key.ResourceNames, key.URLs)
}

// ruleKeys implements sort.Interface
type ruleKeys []ruleKey

func (keys ruleKeys) Len() int           { return len(keys) }
func (keys ruleKeys) Swap(i, j int)      { keys[i], keys[j] = keys[j], keys[i] }
func (keys ruleKeys) Less(i, j int) bool { return keys[i].String() < keys[j].String() }

// knownIrregulars is a helper function to define known irregular kinds and their
// expected formats.
//   - keys   = found values
//   - values = proper values
func knownIrregulars() map[string]string {
	return map[string]string{
		"resourcequota": "resourcequotas",
	}
}

// key normalizes the controller-gen Rule and returns a ruleKey object.
func key(r *rbac.Rule) ruleKey {
	normalize(r)

	return ruleKey{
		Groups:        strings.Join(r.Groups, "&"),
		Resources:     strings.Join(r.Resources, "&"),
		ResourceNames: strings.Join(r.ResourceNames, "&"),
		URLs:          strings.Join(r.URLs, "&"),
	}
}

// addVerbs adds new verbs into a Rule.
// The duplicates in `r.Verbs` will be removed, and then `r.Verbs` will be sorted.
func addVerbs(r *rbac.Rule, verbs []string) {
	r.Verbs = removeDupAndSort(append(r.Verbs, verbs...))
}

// normalize removes duplicates from each field of a Rule, and sorts each field.
func normalize(r *rbac.Rule) {
	r.Groups = removeDupAndSort(r.Groups)
	r.Resources = removeDupAndSort(r.Resources)
	r.ResourceNames = removeDupAndSort(r.ResourceNames)
	r.Verbs = removeDupAndSort(r.Verbs)
	r.URLs = removeDupAndSort(r.URLs)
}

// removeDupAndSort removes duplicates in strs, sorts the items, and returns a
// new slice of strings.
func removeDupAndSort(strs []string) []string {
	set := make(map[string]bool)
	for _, str := range strs {
		if _, ok := set[str]; !ok {
			set[str] = true
		}
	}

	var result []string
	for str := range set {
		result = append(result, str)
	}
	sort.Strings(result)
	return result
}

// GenerateYAML will return the stdout form of rbac objects, given a set of input manifest, in YAML format.
func GenerateYAML(files *manifests.Manifests, options *options.RBACOptions) (string, error) {
	// this is a controller-gen rule, in which we will convert rules from this package into
	rulesByNS := map[string][]*rbac.Rule{}

	for _, manifest := range *files {
		for _, resource := range manifest.ExtractManifests() {
			// decode manifest into unstructured data type
			var manifestObject unstructured.Unstructured

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			if err := runtime.DecodeInto(decoder, []byte(resource), &manifestObject); err != nil {
				return "", fmt.Errorf("%w; unable to decode object in manifest file %s", err, manifest.Filename)
			}

			// determine the rbac rules for this resource
			resourceRules, err := ForResource(&manifestObject)
			if err != nil {
				return "", err
			}

			for _, resourceRule := range *resourceRules {
				rule := &rbac.Rule{
					Groups:    []string{resourceRule.Group},
					Resources: []string{resourceRule.Resource},
					Namespace: manifestObject.GetNamespace(),
					Verbs:     options.Verbs,

					// leave urls empty as we will never have a url derived from a manifest
					URLs: []string{},
				}

				// if we are requesting the resource names, we will also use the resource name as well
				if options.UseResourceNames {
					rule.ResourceNames = []string{manifestObject.GetName()}
				}

				if rulesByNS[manifestObject.GetNamespace()] == nil {
					rulesByNS[manifestObject.GetNamespace()] = []*rbac.Rule{rule}
				} else {
					rulesByNS[manifestObject.GetNamespace()] = append(rulesByNS[manifestObject.GetNamespace()], rule)
				}
			}
		}
	}

	// NormalizeRules merge Rule with the same ruleKey and sort the Rules
	NormalizeRules := func(rules []*rbac.Rule) []rbacv1.PolicyRule {
		ruleMap := make(map[ruleKey]*rbac.Rule)

		// all the Rules having the same ruleKey will be merged into the first Rule
		for _, rule := range rules {
			key := key(rule)
			if _, ok := ruleMap[key]; !ok {
				ruleMap[key] = rule
				continue
			}

			addVerbs(ruleMap[key], rule.Verbs)
		}

		// sort the Rules in rules according to their ruleKeys
		keys := make([]ruleKey, 0, len(ruleMap))
		for key := range ruleMap {
			keys = append(keys, key)
		}

		sort.Sort(ruleKeys(keys))

		var policyRules []rbacv1.PolicyRule
		for _, key := range keys {
			policyRules = append(policyRules, ruleMap[key].ToRule())

		}
		return policyRules
	}

	// collect all the namespaces and sort them
	var namespaces []string
	for ns := range rulesByNS {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	// process the items in rulesByNS by the order specified in `namespaces` to make sure that the Role order is stable
	var roles []client.Object

	for _, ns := range namespaces {
		rules := rulesByNS[ns]
		policyRules := NormalizeRules(rules)
		if len(policyRules) == 0 {
			continue
		}

		if ns == "" {
			roles = append(roles, &rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRole",
					APIVersion: rbacv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: options.RoleName,
				},
				Rules: policyRules,
			})
		} else {
			roles = append(roles, &rbacv1.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: rbacv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      options.RoleName,
					Namespace: ns,
				},
				Rules: policyRules,
			})
		}
	}

	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	var rbacString string

	for i, role := range roles {
		buf := new(bytes.Buffer)

		if err := e.Encode(role, buf); err != nil {
			return "", fmt.Errorf("%w - error encoding role to string", err)
		}

		if i == 0 {
			rbacString = buf.String()
		} else {
			rbacString = fmt.Sprintf("%s---\n%s", rbacString, buf.String())
		}
	}

	return rbacString, nil
}

// GenerateMarkers will return the stdout form of rbac objects as kubebuilder markers.
func GenerateMarkers(files *manifests.Manifests, options *options.RBACOptions) (string, error) {
	var rbacString string

	for _, manifest := range *files {
		for _, resource := range manifest.ExtractManifests() {
			// decode manifest into unstructured data type
			var manifestObject unstructured.Unstructured

			decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

			if err := runtime.DecodeInto(decoder, []byte(resource), &manifestObject); err != nil {
				return "", fmt.Errorf("%w; unable to decode object in manifest file %s", err, manifest.Filename)
			}

			// determine the rbac rules for this resource
			resourceRules, err := ForResource(&manifestObject)
			if err != nil {
				return "", err
			}

			for _, rule := range *resourceRules {
				rule.Verbs = options.Verbs

				if rbacString == "" {
					rbacString = fmt.Sprintf("%s\n", rule.ToMarker())
				} else {
					rbacString = fmt.Sprintf("%s%s\n", rbacString, rule.ToMarker())
				}
			}
		}
	}

	return rbacString, nil
}

// GenerateCode will return the stdout form of rbac objects, given a set of input manifest, in go struct format.
func GenerateCode(files *manifests.Manifests, options *options.RBACOptions) (string, error) {
	var rbacString string

	manifestData, err := GenerateYAML(files, options)
	if err != nil {
		return rbacString, fmt.Errorf("%w - error converting manifests to yaml", err)
	}

	manifest := &manifests.Manifest{Content: []byte(manifestData)}

	for i, resource := range manifest.ExtractManifests() {
		if len(manifest.ExtractManifests()) > 1 {
			options.VariableName = fmt.Sprintf("%s%d", options.VariableName, i)
		}

		asCode, err := code.Generate([]byte(resource), options.VariableName)
		if err != nil {
			return rbacString, fmt.Errorf("%w - error generating code for yaml", err)
		}

		if rbacString == "" {
			rbacString = fmt.Sprintf("%s\n\n", asCode)
		} else {
			rbacString = fmt.Sprintf("%s%s\n", rbacString, asCode)
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
