/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package rbac contain libraries for generating RBAC manifests from RBAC
// markers in Go source files.
//
// The markers take the form:
//
//  +kubebuilder:rbac:groups=<groups>,resources=<resources>,resourceNames=<resource names>,verbs=<verbs>,urls=<non resource urls>
package rbac

import (
	"sort"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// RuleDefinition is a marker for defining RBAC rules.
// Call ToRule on the value to get a Kubernetes RBAC policy rule.
var RuleDefinition = markers.Must(markers.MakeDefinition("kubebuilder:rbac", markers.DescribesPackage, Rule{}))

// +controllertools:marker:generateHelp:category=RBAC

// Rule specifies an RBAC rule to all access to some resources or non-resource URLs.
type Rule struct {
	// Groups specifies the API groups that this rule encompasses.
	Groups []string `marker:",optional"`
	// Resources specifies the API resources that this rule encompasses.
	Resources []string `marker:",optional"`
	// ResourceNames specifies the names of the API resources that this rule encompasses.
	//
	// Create requests cannot be restricted by resourcename, as the object's name
	// is not known at authorization time.
	ResourceNames []string `marker:",optional"`
	// Verbs specifies the (lowercase) kubernetes API verbs that this rule encompasses.
	Verbs []string
	// URL specifies the non-resource URLs that this rule encompasses.
	URLs []string `marker:"urls,optional"`
	// Namespace specifies the scope of the Rule.
	// If not set, the Rule belongs to the generated ClusterRole.
	// If set, the Rule belongs to a Role, whose namespace is specified by this field.
	Namespace string `marker:",optional"`
}

func newSet(strings []string) sets.String {
	if strings == nil { // preserve nils for back-compat with existing code
		return nil
	}

	return sets.NewString(strings...)
}

func (r *Rule) Normalize() *NormalizedRule {
	result := &NormalizedRule{
		Groups:        newSet(r.Groups),
		Resources:     newSet(r.Resources),
		ResourceNames: newSet(r.ResourceNames),
		Verbs:         newSet(r.Verbs),
		URLs:          newSet(r.URLs),
		Namespace:     r.Namespace,
	}

	// simplify Resources and Verbs which both support special "*" to mean 'any'
	// if this is specified, remove all other names
	if result.Resources.Has("*") {
		result.Resources = sets.NewString("*")
	}

	if result.Verbs.Has("*") {
		result.Verbs = sets.NewString("*")
	}

	// fix the group names, since letting people type "core" is nice
	if result.Groups.Has("core") {
		result.Groups.Delete("core")
		result.Groups.Insert("")
	}

	result.GenerateComparisonKey()

	return result
}

type NormalizedRule struct {
	// if two different rules have the same comparison key then they can have their verbs merged
	// this key should comprise all the fields below except Verbs and Namespace (since rules
	// are partitioned by namespace before merging)
	ComparisonKey string

	Namespace     string
	Groups        sets.String
	Resources     sets.String
	ResourceNames sets.String
	URLs          sets.String

	Verbs sets.String
}

// GenerateComparisonKey generates the ComparisonKey
func (nr *NormalizedRule) GenerateComparisonKey() {
	nr.ComparisonKey = strings.Join(
		[]string{
			strings.Join(setToSorted(nr.Groups), "&"),
			strings.Join(setToSorted(nr.Resources), "&"),
			strings.Join(setToSorted(nr.ResourceNames), "&"),
			strings.Join(setToSorted(nr.URLs), "&"),
		},
		" + ")
}

func setToSorted(set sets.String) []string {
	if set == nil { // preserve nils
		return nil
	}

	result := set.UnsortedList()
	sort.Strings(result)
	return result
}

// Subsumes indicates if one rule entirely determines another,
// meaning that the other is unnecessary.
// Remember that Kubernetes RBAC rules are purely additive, there
// are no deny rules.
func (nr *NormalizedRule) Subsumes(other *NormalizedRule) bool {
	// See the code for documentation of these fields: https://github.com/kubernetes/api/blob/v0.23.6/rbac/v1/types.go#L49
	return nr.Namespace == other.Namespace &&
		(nr.Groups.IsSuperset(other.Groups)) &&
		// Resources supports special "*" to mean "any"
		(nr.Resources.Has("*") || nr.Resources.IsSuperset(other.Resources)) &&
		// Empty ResourceNames means "any"
		(len(nr.ResourceNames) == 0 || nr.ResourceNames.IsSuperset(other.ResourceNames)) &&
		nr.URLs.IsSuperset(other.URLs) && // TODO: check? also URLs can have "*" at specific locations
		// Verbs support special "*" to mean "any"
		(nr.Verbs.Has("*") || nr.Verbs.IsSuperset(other.Verbs))
}

// ToRule converts this rule to its Kubernetes API form.
func (nr *NormalizedRule) ToRule() rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups:       setToSorted(nr.Groups),
		Verbs:           setToSorted(nr.Verbs),
		Resources:       setToSorted(nr.Resources),
		ResourceNames:   setToSorted(nr.ResourceNames),
		NonResourceURLs: setToSorted(nr.URLs),
	}
}

// +controllertools:marker:generateHelp

// Generator generates ClusterRole objects.
type Generator struct {
	// RoleName sets the name of the generated ClusterRole.
	RoleName string
}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	if err := into.Register(RuleDefinition); err != nil {
		return err
	}
	into.AddHelp(RuleDefinition, Rule{}.Help())
	return nil
}

// GenerateRoles generate a slice of objs representing either a ClusterRole or a Role object
// The order of the objs in the returned slice is stable and determined by their namespaces.
func GenerateRoles(ctx *genall.GenerationContext, roleName string) ([]interface{}, error) {
	rulesByNS := GroupRulesByNamespace(ctx)

	// collect all the namespaces and sort them
	var namespaces []string
	for ns := range rulesByNS {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	// process the items in rulesByNS by the order specified in `namespaces` to make sure that the Role order is stable
	var objs []interface{}
	for _, ns := range namespaces {
		rules := rulesByNS[ns]
		policyRules := NormalizeRules(rules)
		if len(policyRules) == 0 {
			continue
		}
		if ns == "" {
			objs = append(objs, rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRole",
					APIVersion: rbacv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: roleName,
				},
				Rules: policyRules,
			})
		} else {
			objs = append(objs, rbacv1.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: rbacv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: ns,
				},
				Rules: policyRules,
			})
		}
	}

	return objs, nil
}

func GroupRulesByNamespace(ctx *genall.GenerationContext) map[string][]*Rule {
	rulesByNS := make(map[string][]*Rule)
	for _, root := range ctx.Roots {
		markerSet, err := markers.PackageMarkers(ctx.Collector, root)
		if err != nil {
			root.AddError(err)
		}

		// group RBAC markers by namespace
		for _, markerValue := range markerSet[RuleDefinition.Name] {
			rule := markerValue.(Rule)
			namespace := rule.Namespace
			rulesByNS[namespace] = append(rulesByNS[namespace], &rule)
		}
	}

	return rulesByNS
}

// insertRule inserts a rule into a destination slice, deduplicating via the
// Subsumes function or merging verbs if appropriate
func insertRule(dest []*NormalizedRule, it *NormalizedRule) []*NormalizedRule {
	// this is not going to be very fast but the set of rules should always be small
	mergeWith := -1
	for ix, other := range dest {
		if other.Subsumes(it) {
			// not needed; another rule handles this case
			return dest
		}

		if it.Subsumes(other) {
			// rebuild whole list:
			// the 'it' rule subsumes the 'other' rule;
			// but it might also subsume other rules in the list, so we
			// need to go through and check them as well.
			//
			// redoing the insertion with `it` first handles this:
			result := []*NormalizedRule{it}
			for _, d := range dest {
				result = insertRule(result, d)
			}

			return result
		}

		if it.ComparisonKey == other.ComparisonKey {
			// match the same things, can merge their
			// verbs (if no better match)
			mergeWith = ix
		}
	}

	if mergeWith >= 0 {
		// we found one to merge with
		insertAll(dest[mergeWith].Verbs, it.Verbs)
		return dest
	}

	// otherwise, insert it
	return append(dest, it)
}

func insertAll(set sets.String, other sets.String) {
	for value := range other {
		set.Insert(value)
	}
}

// Sorts the rules for deterministic output:
type ruleSorter []*NormalizedRule

// ruleSorter implements sort.Interface
var _ sort.Interface = ruleSorter{}

func (keys ruleSorter) Len() int           { return len(keys) }
func (keys ruleSorter) Swap(i, j int)      { keys[i], keys[j] = keys[j], keys[i] }
func (keys ruleSorter) Less(i, j int) bool { return keys[i].ComparisonKey < keys[j].ComparisonKey }

// NormalizeRules merges Rules that can be merged, and sorts the Rules
func NormalizeRules(rules []*Rule) []rbacv1.PolicyRule {
	var simplified []*NormalizedRule
	for _, rule := range rules {
		simplified = insertRule(simplified, rule.Normalize())
	}

	sort.Sort(ruleSorter(simplified))

	result := make([]rbacv1.PolicyRule, len(simplified))
	for i := range simplified {
		result[i] = simplified[i].ToRule()
	}

	return result
}

func (g Generator) Generate(ctx *genall.GenerationContext) error {
	objs, err := GenerateRoles(ctx, g.RoleName)
	if err != nil {
		return err
	}

	if len(objs) == 0 {
		return nil
	}

	return ctx.WriteYAML("role.yaml", objs)
}
