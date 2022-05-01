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

	"github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
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

func Test_SubsumesIsReflexive(t *testing.T) {
	g := gomega.NewWithT(t)

	for _, nr := range nrules {
		g.Expect(nr.Subsumes(nr)).To(gomega.BeTrue())
	}
}

func Test_SubsumesIsOneWay(t *testing.T) {
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

func Test_StarInGroupsMeansAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	rule1 := Rule{
		Groups:    []string{"cluster.x-k8s.io"},
		Resources: []string{"machinedeployments"},
		Verbs:     []string{"update", "patch"},
	}

	rule2 := rule1
	rule2.Groups = []string{"*"}

	g.Expect(rule2.Normalize().Subsumes(rule1.Normalize())).To(gomega.BeTrue())
}

func Test_StarInResourcesMeansAnything(t *testing.T) {
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

func Test_StarInResourceNamesMeansAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	rule1 := Rule{
		Groups:        []string{"cluster.x-k8s.io"},
		Resources:     []string{"machinedeployments"},
		ResourceNames: []string{"resourcename"},
		Verbs:         []string{"update", "patch"},
	}

	rule2 := rule1
	rule2.ResourceNames = []string{"*"}

	g.Expect(rule2.Normalize().Subsumes(rule1.Normalize())).To(gomega.BeTrue())
}
