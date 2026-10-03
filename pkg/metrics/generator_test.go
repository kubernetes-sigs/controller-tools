/*
Copyright 2026 The Kubernetes Authors.

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

package metrics

import (
	"slices"
	"strings"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_sortedResources(t *testing.T) {
	newResource := func(version string, names ...string) *model.Resource {
		r := &model.Resource{GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: version, Kind: "Foo"}}
		for _, name := range names {
			r.Metrics = append(r.Metrics, model.Generator{Name: name, Help: name + "-" + version})
		}
		return r
	}

	resourceStates := map[crd.TypeIdent]*model.Resource{
		{Name: "v2"}:      newResource("v2", "b", "a"),
		{Name: "v1"}:      newResource("v1", "same", "same"),
		{Name: "nil"}:     nil,
		{Name: "metrics"}: newResource("v3"),
	}
	// The stable sort must keep the order of the metrics with the same name.
	resourceStates[crd.TypeIdent{Name: "v1"}].Metrics[0].Help = "first"
	resourceStates[crd.TypeIdent{Name: "v1"}].Metrics[1].Help = "second"

	// Iterating over a map is random, so run multiple times to make a flaky ordering visible.
	for range 20 {
		got := sortedResources(resourceStates)

		versions := make([]string, 0, len(got))
		for _, r := range got {
			versions = append(versions, r.GroupVersionKind.Version)
		}
		if want := []string{"v1", "v2"}; !slices.Equal(versions, want) {
			t.Fatalf("sortedResources() versions = %v, want %v (resources without metrics and nil must be dropped, sorted by GVK)", versions, want)
		}

		if got[0].Metrics[0].Help != "first" || got[0].Metrics[1].Help != "second" {
			t.Errorf("sortedResources() changed the order of metrics with the same name: %v", got[0].Metrics)
		}
		if got[1].Metrics[0].Name != "a" || got[1].Metrics[1].Name != "b" {
			t.Errorf("sortedResources() did not sort the metrics by name: %v", got[1].Metrics)
		}
	}
}

func Test_Generate_invalidTargetFailsBeforeParsing(t *testing.T) {
	// A nil GenerationContext would panic if the generator started parsing.
	err := Generator{Experimental: true, Target: "foo"}.Generate(nil)
	if err == nil || !strings.Contains(err.Error(), `unsupported metrics target "foo"`) {
		t.Fatalf("Generate() error = %v, want an unsupported target error", err)
	}
}

func Test_Generate_requiresExperimental(t *testing.T) {
	err := Generator{}.Generate(nil)
	if err == nil || !strings.Contains(err.Error(), "experimental") {
		t.Fatalf("Generate() error = %v, want an error about the experimental flag", err)
	}
}

func Test_addResources_validatesForAllTargets(t *testing.T) {
	// An unsupported metric: labels for the entries of a map in a gauge.
	resources := []model.Resource{{
		GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Foo"},
		Metrics: []model.Generator{{
			Name: "foo", Help: "Foo.", Type: model.MetricTypeGauge,
			Path: model.Path{"spec", "m"}, PathKind: model.PathKindMap, KeyLabel: "key",
		}},
	}}

	for _, target := range []string{targetResourceStateMetrics, targetKubeStateMetrics} {
		t.Run(target, func(t *testing.T) {
			builder, err := Generator{Target: target}.newOutputBuilder()
			if err != nil {
				t.Fatal(err)
			}
			err = addResources(builder, resources)
			if err == nil || !strings.Contains(err.Error(), "a Gauge metric on a map doesn't support labels") {
				t.Fatalf("addResources() error = %v, want an error about labels on a map", err)
			}
		})
	}
}
