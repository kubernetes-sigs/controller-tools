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

package rbac

import (
	"testing"
	"testing/quick"

	"github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

//From: https://github.com/kubernetes-sigs/controller-tools/issues/612, but altered slightly (#3 & #4 made distinct)
//- kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments;machinedeployments/status;machinedeployments/finalizers,verbs=get;list;watch;create;update;patch;delete
//- kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments;machinedeployments/finalizers,verbs=get;list;watch;update;patch
//- kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments,verbs=get;list;watch
//- kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments,verbs=update;patch
var rules = []*Rule{
	{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments", "machinedeployments/status", "machinedeployments/finalizers"},
		Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
	},
	{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments", "machinedeployments/finalizers"},
		Verbs:     []string{"get", "list", "watch", "update", "patch"},
	},
	{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments"},
		Verbs:     []string{"update", "patch"},
	},
}

var nrules = normalizeRules(rules)

func normalizeRules(rules []*Rule) []*NormalizedRule {
	result := make([]*NormalizedRule, len(rules))
	for ix := range rules {
		result[ix] = rules[ix].Normalize()
	}

	return result
}

func Test_Subsume_Simple(t *testing.T) {
	g := gomega.NewWithT(t)
	// subsumes all others:
	for _, nrule := range nrules {
		g.Expect(nrules[0].Subsumes(nrule)).To(gomega.BeTrue())
	}

	g.Expect(nrules[1].Subsumes(nrules[2])).To(gomega.BeTrue())
	g.Expect(nrules[1].Subsumes(nrules[3])).To(gomega.BeTrue())

	// distinct
	g.Expect(nrules[3].Subsumes(nrules[2])).To(gomega.BeFalse())
	g.Expect(nrules[2].Subsumes(nrules[3])).To(gomega.BeFalse())
}

func Test_SubsumesIsReflexive_CannedExamples(t *testing.T) {
	g := gomega.NewWithT(t)

	for _, nr := range nrules {
		g.Expect(nr.Subsumes(nr)).To(gomega.BeTrue())
	}
}

func Test_SubsumesIsReflexive_PropertyTest(t *testing.T) {
	f := func(groups, verbs, resources, resourceNames []string) bool {
		rule := Rule{Groups: groups, Verbs: verbs, Resources: resources, ResourceNames: resourceNames}
		normed := rule.Normalize()

		return normed.Subsumes(normed)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func Test_SubsumesIsOneWay(t *testing.T) {
	// note that this does not hold for any arbitrary rules
	// given that Subsumes is Reflexive (as asserted by previous test)
	// but it holds for all the examples we have listed above:

	g := gomega.NewWithT(t)

	for i := range nrules {
		for j := range nrules {
			if i == j {
				continue
			}

			ruleI := nrules[i]
			ruleJ := nrules[j]

			if ruleI.Subsumes(ruleJ) {
				g.Expect(ruleJ.Subsumes(ruleI)).To(gomega.BeFalse())
			}
		}
	}
}

func Test_Simplification(t *testing.T) {
	g := gomega.NewWithT(t)

	g.Expect(NormalizeRules(rules)).To(gomega.Equal([]rbacv1.PolicyRule{
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"cluster.x-k8s.io"},
			Resources: []string{
				"machinedeployments",
				"machinedeployments/finalizers",
				"machinedeployments/status",
			},
		},
	}))
}

func Test_Simplification_OrderIrrelevant(t *testing.T) {
	g := gomega.NewWithT(t)

	// reversed order
	rules2 := []*Rule{rules[3], rules[2], rules[1], rules[0]}

	// expect same outcome as previous test
	g.Expect(NormalizeRules(rules2)).To(gomega.Equal([]rbacv1.PolicyRule{
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"cluster.x-k8s.io"},
			Resources: []string{
				"machinedeployments",
				"machinedeployments/finalizers",
				"machinedeployments/status",
			},
		},
	}))
}

func Test_Simplification_Merge(t *testing.T) {
	g := gomega.NewWithT(t)

	// these two rules should merge into one
	g.Expect(NormalizeRules([]*Rule{rules[2], rules[3]})).To(gomega.Equal([]rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "patch", "update", "watch"},
			APIGroups: []string{"cluster.x-k8s.io"},
			Resources: []string{"machinedeployments"},
		},
	}))
}

func Test_StarInGroupsDoesNotHaveMeaning(t *testing.T) {
	g := gomega.NewWithT(t)

	rule1 := Rule{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments"},
		Verbs:     []string{"update", "patch"},
	}

	rule2 := rule1
	rule2.Groups = []string{"*"}

	g.Expect(rule2.Normalize().Subsumes(rule1.Normalize())).To(gomega.BeFalse())
}

func Test_StarInResourcesMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	rule1 := Rule{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments"},
		Verbs:     []string{"update", "patch"},
	}

	rule2 := rule1
	rule2.Resources = []string{"*"}

	g.Expect(rule2.Normalize().Subsumes(rule1.Normalize())).To(gomega.BeTrue())
}

func Test_StarInResourceNamesDoesNotHaveMeaning(t *testing.T) {
	g := gomega.NewWithT(t)

	rule1 := Rule{
		Groups:        []string{"cluster.x-k8s.io"},
		Resources:     []string{"machinedeployments"},
		ResourceNames: []string{"resourcename"},
		Verbs:         []string{"update", "patch"},
	}

	rule2 := rule1
	rule2.ResourceNames = []string{"*"}

	g.Expect(rule2.Normalize().Subsumes(rule1.Normalize())).To(gomega.BeFalse())
}

func Test_StarInVerbsMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	rule1 := Rule{
		Groups:        []string{"cluster.x-k8s.io"},
		Resources:     []string{"machinedeployments"},
		ResourceNames: []string{"resourcename"},
		Verbs:         []string{"update", "patch"},
	}

	rule2 := rule1
	rule2.Verbs = []string{"*"}

	g.Expect(rule2.Normalize().Subsumes(rule1.Normalize())).To(gomega.BeTrue())
}

func Test_StarInResourcesSimplifies(t *testing.T) {
	// star is special in Resources

	g := gomega.NewWithT(t)

	rule := Rule{
		Resources: []string{"some", "*", "names"},
	}

	normalized := rule.Normalize()
	g.Expect(normalized.Resources).To(gomega.Equal(sets.NewString("*")))
}

func Test_StarInVerbsSimplifies(t *testing.T) {
	// star is special in Verbs

	g := gomega.NewWithT(t)

	rule := Rule{
		Verbs: []string{"some", "*", "names"},
	}

	normalized := rule.Normalize()
	g.Expect(normalized.Verbs).To(gomega.Equal(sets.NewString("*")))
}

func Test_StarInGroupsDoesNotSimplify(t *testing.T) {
	// star is not special in Groups

	g := gomega.NewWithT(t)

	rule := Rule{
		Groups: []string{"some", "*", "things"},
	}

	normalized := rule.Normalize()
	g.Expect(normalized.Groups).To(gomega.Equal(sets.NewString("some", "*", "things")))
}

func Test_StarInResourceNamesDoesNotSimplify(t *testing.T) {
	// star is not special in Resource Names

	g := gomega.NewWithT(t)

	rule := Rule{
		ResourceNames: []string{"some", "*", "things"},
	}

	normalized := rule.Normalize()
	g.Expect(normalized.ResourceNames).To(gomega.Equal(sets.NewString("some", "*", "things")))
}

func Test_EmptyResourceNameListMatchesAnything(t *testing.T) {
	f := func(groups, verbs, resources, resourceNames []string) bool {
		rule := Rule{Groups: groups, Verbs: verbs, Resources: resources, ResourceNames: resourceNames}

		biggerRule := rule
		biggerRule.ResourceNames = nil

		return biggerRule.Normalize().Subsumes(rule.Normalize())
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func Test_EmptyGroupListDoesNotMatchEverything(t *testing.T) {
	f := func(groups, verbs, resources, resourceNames []string) bool {
		rule := Rule{Groups: groups, Verbs: verbs, Resources: resources, ResourceNames: resourceNames}

		biggerRule := rule
		biggerRule.Groups = nil

		// subsumes only if the groups list was originally empty
		return biggerRule.Normalize().Subsumes(rule.Normalize()) == (len(rule.Groups) == 0)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func Test_EmptyResourcesListDoesNotMatchEverything(t *testing.T) {
	f := func(groups, verbs, resources, resourceNames []string) bool {
		rule := Rule{Groups: groups, Verbs: verbs, Resources: resources, ResourceNames: resourceNames}

		biggerRule := rule
		biggerRule.Resources = nil

		// subsumes only if the resources list was originally empty
		return biggerRule.Normalize().Subsumes(rule.Normalize()) == (len(rule.Resources) == 0)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func Test_EmptyVerbsourcesListDoesNotMatchEverything(t *testing.T) {
	f := func(groups, verbs, resources, resourceNames []string) bool {
		rule := Rule{Groups: groups, Verbs: verbs, Resources: resources, ResourceNames: resourceNames}

		biggerRule := rule
		biggerRule.Verbs = nil

		// subsumes only if the verbs list was originally empty
		return biggerRule.Normalize().Subsumes(rule.Normalize()) == (len(rule.Verbs) == 0)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
