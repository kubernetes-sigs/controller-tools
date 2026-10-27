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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_sortedResources(t *testing.T) {
	gvk := func(version string) model.GroupVersionKind {
		return model.GroupVersionKind{Group: "example.io", Version: version, Kind: "Foo"}
	}

	resourceStates := map[crd.TypeIdent]*model.Resource{
		{Name: "v2"}: {GroupVersionKind: gvk("v2"), Metrics: []model.Generator{{Name: "b"}, {Name: "a"}}},
		// The metrics with the same name keep their order, as the sort is stable.
		{Name: "v1"}: {GroupVersionKind: gvk("v1"), Metrics: []model.Generator{{Name: "same", Help: "first"}, {Name: "same", Help: "second"}}},
		// Resources which are nil or don't have metrics are dropped.
		{Name: "nil"}:     nil,
		{Name: "metrics"}: {GroupVersionKind: gvk("v3")},
	}

	// The resources are sorted by GVK and their metrics by name.
	want := []model.Resource{
		{GroupVersionKind: gvk("v1"), Metrics: []model.Generator{{Name: "same", Help: "first"}, {Name: "same", Help: "second"}}},
		{GroupVersionKind: gvk("v2"), Metrics: []model.Generator{{Name: "a"}, {Name: "b"}}},
	}

	// Iterating over a map is random, so run multiple times to make a flaky ordering visible.
	for range 20 {
		if diff := cmp.Diff(want, sortedResources(resourceStates)); diff != "" {
			t.Fatalf("sortedResources() (-want,+got): %s", diff)
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
