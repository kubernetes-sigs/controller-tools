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

package rsm

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

// equateEmpty treats nil and empty slices as equal, to not have to set empty labels in the expected metrics.
var equateEmpty = cmpopts.EquateEmpty()

func Test_toStore(t *testing.T) {
	prefix := "foo"
	resource := model.Resource{
		// The name prefix is ignored, the prefix of the families is derived from the kind.
		NamePrefix:       &prefix,
		GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
		Labels: []model.Label{
			{Name: "cluster", Path: model.Path{"metadata", "labels", "cluster"}},
		},
		Metrics: []model.Generator{
			{
				IgnoreMissing: true, Name: "condition", Help: "A condition.", Type: model.MetricTypeStateSet,
				Path: model.Path{"status", "conditions"}, PathKind: model.PathKindArray,
				Labels:    []model.Label{{Name: "type", Path: model.Path{"type"}}},
				List:      []string{"True", "False"},
				LabelName: "status",
				Value:     model.Path{"status"},
			},
			{
				IgnoreMissing: true, Name: "transition", Help: "Transition time.", Type: model.MetricTypeGauge,
				Path: model.Path{"status", "conditions"}, PathKind: model.PathKindArray,
				Value: model.Path{"lastTransitionTime"}, ValueKind: model.ValueKindTimestamp,
			},
		},
	}

	const (
		conditions = `has(o.status) && has(o.status.conditions) ? o.status.conditions.filter(v, has(v.status))`
		// The labels of the resource are added to every metric instead of the store, see toStore.
		clusterLabel = `has(o.metadata) && has(o.metadata.labels) && has(o.metadata.labels.cluster) ? string(o.metadata.labels.cluster) : ""`
		typeLabel    = conditions + `.map(v, has(v.type) ? string(v.type) : "") : []`
	)
	want := Store{
		Group: "example.io", Version: "v1", Kind: "Policy", Resource: "policies", Resolver: resolverCEL,
		Families: []Family{
			{
				Name: "policy_condition", Help: "A condition.",
				Metrics: []Metric{
					{
						Labels: []Label{{Name: "type", Value: typeLabel}, {Name: "status", Value: `"True"`}, {Name: "cluster", Value: clusterLabel}},
						Value:  conditions + `.map(v, int(v.status == "True" ? 1 : 0)) : []`,
					},
					{
						Labels: []Label{{Name: "type", Value: typeLabel}, {Name: "status", Value: `"False"`}, {Name: "cluster", Value: clusterLabel}},
						Value:  conditions + `.map(v, int(v.status == "False" ? 1 : 0)) : []`,
					},
				},
			},
			{
				Name: "policy_transition", Help: "Transition time.",
				Metrics: []Metric{{
					Labels: []Label{{Name: "cluster", Value: clusterLabel}},
					Value:  `has(o.status) && has(o.status.conditions) ? o.status.conditions.filter(v, has(v.lastTransitionTime)).map(v, unixSeconds(v.lastTransitionTime)) : []`,
				}},
			},
		},
	}

	if diff := cmp.Diff(want, toStore(resource)); diff != "" {
		t.Errorf("toStore() (-want,+got): %s", diff)
	}
}

// Test_toCEL_withoutIgnoreMissing tests the expressions without the checks for missing fields.
func Test_toCEL_withoutIgnoreMissing(t *testing.T) {
	t.Run("gauge", func(t *testing.T) {
		tests := []struct {
			name   string
			metric model.Generator
			want   Metric
		}{
			{
				name:   "scalar",
				metric: model.Generator{Path: model.Path{"spec", "replicas"}, PathKind: model.PathKindScalar},
				want:   Metric{Value: `o.spec.replicas`},
			},
			{
				// missingAsZero needs the check independent of ignoreMissing.
				name:   "missing as zero",
				metric: model.Generator{Path: model.Path{"spec", "paused"}, PathKind: model.PathKindScalar, MissingAsZero: true},
				want:   Metric{Value: `has(o.spec) && has(o.spec.paused) ? o.spec.paused : 0`},
			},
			{
				name:   "list",
				metric: model.Generator{Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray},
				want:   Metric{Value: `o.status.items.map(v, v.size)`},
			},
			{
				name:   "list elements are zero with missing as zero",
				metric: model.Generator{Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray, MissingAsZero: true},
				want:   Metric{Value: `o.status.items.map(v, has(v.size) ? v.size : 0)`},
			},
			{
				name: "list labels fail with the value",
				metric: model.Generator{
					Path: model.Path{"status", "items"}, PathKind: model.PathKindArray, Value: model.Path{"size"},
					Labels: []model.Label{{Name: "id", Path: model.Path{"id"}}},
				},
				want: Metric{
					Labels: []Label{{Name: "id", Value: `o.status.items.map(v, string(v.id))`}},
					Value:  `o.status.items.map(v, v.size)`,
				},
			},
			{
				name:   "map",
				metric: model.Generator{Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap},
				want:   Metric{Value: `o.status.pools.map(k, o.status.pools[k].size)`},
			},
			{
				name:   "timestamp",
				metric: model.Generator{Path: model.Path{"status", "at"}, PathKind: model.PathKindScalar, ValueKind: model.ValueKindTimestamp},
				want:   Metric{Value: `unixSeconds(o.status.at)`},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.metric.IgnoreMissing = false
				if diff := cmp.Diff(tt.want, gaugeToCEL(tt.metric), equateEmpty); diff != "" {
					t.Errorf("gaugeToCEL() (-want,+got): %s", diff)
				}
			})
		}
	})

	t.Run("info", func(t *testing.T) {
		tests := []struct {
			name   string
			metric model.Generator
			want   Metric
		}{
			{
				name:   "no path",
				metric: model.Generator{PathKind: model.PathKindScalar},
				want:   Metric{Value: `1`},
			},
			{
				// The value has to fail like the labels if the object is missing.
				name: "object",
				metric: model.Generator{
					Path: model.Path{"status", "nodeRef"}, PathKind: model.PathKindObject,
					Labels: []model.Label{{Name: "node_name", Path: model.Path{"name"}}},
				},
				want: Metric{
					Labels: []Label{{Name: "node_name", Value: `string(o.status.nodeRef.name)`}},
					Value:  `size([o.status.nodeRef])`,
				},
			},
			{
				name: "list",
				metric: model.Generator{
					Path: model.Path{"metadata", "ownerReferences"}, PathKind: model.PathKindArray,
					Labels: []model.Label{{Name: "owner_kind", Path: model.Path{"kind"}}},
				},
				want: Metric{
					Labels: []Label{{Name: "owner_kind", Value: `o.metadata.ownerReferences.map(v, string(v.kind))`}},
					Value:  `o.metadata.ownerReferences.map(v, 1)`,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.metric.IgnoreMissing = false
				if diff := cmp.Diff(tt.want, infoToCEL(tt.metric), equateEmpty); diff != "" {
					t.Errorf("infoToCEL() (-want,+got): %s", diff)
				}
			})
		}
	})

	t.Run("stateset", func(t *testing.T) {
		tests := []struct {
			name   string
			metric model.Generator
			want   []Metric
		}{
			{
				name: "scalar",
				metric: model.Generator{
					Path: model.Path{"status", "phase"}, PathKind: model.PathKindScalar,
					List: []string{"Ready"}, LabelName: "phase",
				},
				want: []Metric{{
					Labels: []Label{{Name: "phase", Value: `"Ready"`}},
					Value:  `int(o.status.phase == "Ready" ? 1 : 0)`,
				}},
			},
			{
				name: "list with labels",
				metric: model.Generator{
					Path: model.Path{"status", "conditions"}, PathKind: model.PathKindArray,
					Value: model.Path{"status"}, List: []string{"True"}, LabelName: "status",
					Labels: []model.Label{{Name: "type", Path: model.Path{"type"}}},
				},
				want: []Metric{{
					Labels: []Label{
						{Name: "type", Value: `o.status.conditions.map(v, string(v.type))`},
						{Name: "status", Value: `"True"`},
					},
					Value: `o.status.conditions.map(v, int(v.status == "True" ? 1 : 0))`,
				}},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.metric.IgnoreMissing = false
				if diff := cmp.Diff(tt.want, stateSetToCEL(tt.metric), equateEmpty); diff != "" {
					t.Errorf("stateSetToCEL() (-want,+got): %s", diff)
				}
			})
		}
	})
}

func Test_labelsToCEL_isString(t *testing.T) {
	labels := []model.Label{
		{Name: "text", Path: model.Path{"text"}, IsString: true},
		{Name: "flag", Path: model.Path{"flag"}},
	}
	tests := []struct {
		name          string
		ignoreMissing bool
		path          model.Path
		kind          model.PathKind
		want          []Label
	}{
		{
			name: "scalar", path: model.Path{"spec"}, kind: model.PathKindObject,
			want: []Label{
				{Name: "text", Value: `o.spec.text`},
				{Name: "flag", Value: `string(o.spec.flag)`},
			},
		},
		{
			name: "scalar with ignoreMissing", ignoreMissing: true, path: model.Path{"spec"}, kind: model.PathKindObject,
			want: []Label{
				{Name: "text", Value: `has(o.spec) && has(o.spec.text) ? o.spec.text : ""`},
				{Name: "flag", Value: `has(o.spec) && has(o.spec.flag) ? string(o.spec.flag) : ""`},
			},
		},
		{
			name: "list", path: model.Path{"status", "items"}, kind: model.PathKindArray,
			want: []Label{
				{Name: "text", Value: `o.status.items.map(v, v.text)`},
				{Name: "flag", Value: `o.status.items.map(v, string(v.flag))`},
			},
		},
		{
			name: "list with ignoreMissing", ignoreMissing: true, path: model.Path{"status", "items"}, kind: model.PathKindArray,
			want: []Label{
				{Name: "text", Value: `has(o.status) && has(o.status.items) ? o.status.items.map(v, has(v.text) ? v.text : "") : []`},
				{Name: "flag", Value: `has(o.status) && has(o.status.items) ? o.status.items.map(v, has(v.flag) ? string(v.flag) : "") : []`},
			},
		},
		{
			name: "map", path: model.Path{"status", "pools"}, kind: model.PathKindMap,
			want: []Label{
				{Name: "text", Value: `o.status.pools.map(k, o.status.pools[k].text)`},
				{Name: "flag", Value: `o.status.pools.map(k, string(o.status.pools[k].flag))`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, labelsToCEL(labels, tt.path, tt.kind, "", tt.ignoreMissing)); diff != "" {
				t.Errorf("labelsToCEL() (-want,+got): %s", diff)
			}
		})
	}
}

func Test_stateSetToCEL(t *testing.T) {
	tests := []struct {
		name   string
		metric model.Generator
		want   []Metric
	}{
		{
			name: "scalar",
			metric: model.Generator{IgnoreMissing: true,
				Path: model.Path{"status", "phase"}, PathKind: model.PathKindScalar,
				List: []string{"Ready", "Failed"}, LabelName: "phase",
			},
			want: []Metric{
				{
					Labels: []Label{{Name: "phase", Value: `"Ready"`}},
					Value:  `has(o.status) && has(o.status.phase) ? [int(o.status.phase == "Ready" ? 1 : 0)] : []`,
				},
				{
					Labels: []Label{{Name: "phase", Value: `"Failed"`}},
					Value:  `has(o.status) && has(o.status.phase) ? [int(o.status.phase == "Failed" ? 1 : 0)] : []`,
				},
			},
		},
		{
			name: "list with labels",
			metric: model.Generator{IgnoreMissing: true,
				Path: model.Path{"status", "conditions"}, PathKind: model.PathKindArray,
				Value: model.Path{"status"}, List: []string{"True"}, LabelName: "status",
				Labels: []model.Label{{Name: "type", Path: model.Path{"type"}}},
			},
			want: []Metric{{
				// The labels use the same filter as the value, as the entries of the lists have to be aligned.
				Labels: []Label{
					{Name: "type", Value: `has(o.status) && has(o.status.conditions) ? o.status.conditions.filter(v, has(v.status)).map(v, has(v.type) ? string(v.type) : "") : []`},
					{Name: "status", Value: `"True"`},
				},
				Value: `has(o.status) && has(o.status.conditions) ? o.status.conditions.filter(v, has(v.status)).map(v, int(v.status == "True" ? 1 : 0)) : []`,
			}},
		},
		{
			name: "map",
			metric: model.Generator{IgnoreMissing: true,
				Path: model.Path{"status", "pools"}, PathKind: model.PathKindMap,
				Value: model.Path{"phase"}, List: []string{"Ready"}, LabelName: "phase",
			},
			want: []Metric{{
				Labels: []Label{{Name: "phase", Value: `"Ready"`}},
				Value:  `has(o.status) && has(o.status.pools) ? o.status.pools.filter(k, has(o.status.pools[k].phase)).map(k, int(o.status.pools[k].phase == "Ready" ? 1 : 0)) : []`,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, stateSetToCEL(tt.metric), equateEmpty); diff != "" {
				t.Errorf("stateSetToCEL() (-want,+got): %s", diff)
			}
		})
	}
}

func Test_infoToCEL(t *testing.T) {
	tests := []struct {
		name   string
		metric model.Generator
		want   Metric
	}{
		{
			name: "object",
			metric: model.Generator{IgnoreMissing: true,
				Path: model.Path{"status", "nodeRef"}, PathKind: model.PathKindObject,
				Labels: []model.Label{{Name: "node_name", Path: model.Path{"name"}}},
			},
			want: Metric{
				Labels: []Label{{Name: "node_name", Value: `has(o.status) && has(o.status.nodeRef) && has(o.status.nodeRef.name) ? string(o.status.nodeRef.name) : ""`}},
				Value:  `has(o.status) && has(o.status.nodeRef) ? [1] : []`,
			},
		},
		{
			name: "list",
			metric: model.Generator{IgnoreMissing: true,
				Path: model.Path{"metadata", "ownerReferences"}, PathKind: model.PathKindArray,
				Labels: []model.Label{{Name: "owner_kind", Path: model.Path{"kind"}}},
			},
			want: Metric{
				Labels: []Label{{Name: "owner_kind", Value: `has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, has(v.kind) ? string(v.kind) : "") : []`}},
				Value:  `has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, 1) : []`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, infoToCEL(tt.metric), equateEmpty); diff != "" {
				t.Errorf("infoToCEL() (-want,+got): %s", diff)
			}
		})
	}
}

func Test_gaugeToCEL(t *testing.T) {
	tests := []struct {
		name   string
		metric model.Generator
		want   Metric
	}{
		{
			name:   "scalar",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"spec", "replicas"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.spec) && has(o.spec.replicas) ? [o.spec.replicas] : []`},
		},
		{
			name:   "missing as zero",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"spec", "paused"}, PathKind: model.PathKindScalar, MissingAsZero: true},
			want:   Metric{Value: `has(o.spec) && has(o.spec.paused) ? o.spec.paused : 0`},
		},
		{
			name:   "value path of an object",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "ref"}, Value: model.Path{"count"}, PathKind: model.PathKindObject},
			want:   Metric{Value: `has(o.status) && has(o.status.ref) && has(o.status.ref.count) ? [o.status.ref.count] : []`},
		},
		{
			name:   "list elements without the value are left out",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray},
			want:   Metric{Value: `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, v.size) : []`},
		},
		{
			name:   "list elements without the value are zero with missing as zero",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray, MissingAsZero: true},
			want:   Metric{Value: `has(o.status) && has(o.status.items) ? o.status.items.map(v, has(v.size) ? v.size : 0) : []`},
		},
		{
			name:   "list of values",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "sizes"}, PathKind: model.PathKindArray},
			want:   Metric{Value: `has(o.status) && has(o.status.sizes) ? o.status.sizes.map(v, v) : []`},
		},
		{
			// The labels use the same filter as the value, as the entries of the lists have to be aligned.
			name: "list labels use the filter of the value",
			metric: model.Generator{IgnoreMissing: true,
				Path: model.Path{"status", "items"}, PathKind: model.PathKindArray, Value: model.Path{"size"},
				Labels: []model.Label{{Name: "id", Path: model.Path{"id"}}},
			},
			want: Metric{
				Labels: []Label{{Name: "id", Value: `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, has(v.id) ? string(v.id) : "") : []`}},
				Value:  `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, v.size) : []`,
			},
		},
		{
			name:   "map entries without the value are left out",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap},
			want:   Metric{Value: `has(o.status) && has(o.status.pools) ? o.status.pools.filter(k, has(o.status.pools[k].size)).map(k, o.status.pools[k].size) : []`},
		},
		{
			name:   "map entries without the value are zero with missing as zero",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap, MissingAsZero: true},
			want:   Metric{Value: `has(o.status) && has(o.status.pools) ? o.status.pools.map(k, has(o.status.pools[k].size) ? o.status.pools[k].size : 0) : []`},
		},
		{
			name:   "map of values",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "counts"}, PathKind: model.PathKindMap},
			want:   Metric{Value: `has(o.status) && has(o.status.counts) ? o.status.counts.map(k, o.status.counts[k]) : []`},
		},
		{
			name:   "map with a value path",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"spec", "pools"}, PathKind: model.PathKindMap, Value: model.Path{"size"}},
			want: Metric{
				Value: `has(o.spec) && has(o.spec.pools) ? o.spec.pools.filter(k, has(o.spec.pools[k].size)).map(k, o.spec.pools[k].size) : []`,
			},
		},
		{
			name:   "timestamp type in a map",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "at"}, Value: model.Path{"time"}, PathKind: model.PathKindMap, ValueKind: model.ValueKindTimestamp},
			want:   Metric{Value: `has(o.status) && has(o.status.at) ? o.status.at.filter(k, has(o.status.at[k].time)).map(k, unixSeconds(o.status.at[k].time)) : []`},
		},
		{
			name:   "timestamp type",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "at"}, PathKind: model.PathKindScalar, ValueKind: model.ValueKindTimestamp},
			want:   Metric{Value: `has(o.status) && has(o.status.at) ? [unixSeconds(o.status.at)] : []`},
		},
		{
			// The conversion would fail for numeric fields like these.
			name:   "no timestamp conversion based on the field name",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"status", "uptime"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.status) && has(o.status.uptime) ? [o.status.uptime] : []`},
		},
		{
			name:   "field names which are CEL keywords",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"spec", "in", "null"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.spec) && "in" in o.spec && "null" in o.spec["in"] ? [o.spec["in"]["null"]] : []`},
		},
		{
			name:   "non identifier field name",
			metric: model.Generator{IgnoreMissing: true, Path: model.Path{"metadata", "labels", "cluster.x-k8s.io/name"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.metadata) && has(o.metadata.labels) && "cluster.x-k8s.io/name" in o.metadata.labels ? [o.metadata.labels["cluster.x-k8s.io/name"]] : []`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, gaugeToCEL(tt.metric), equateEmpty); diff != "" {
				t.Errorf("gaugeToCEL() (-want,+got): %s", diff)
			}
		})
	}
}

func Test_guardedPathCEL_neverSelectsKeywords(t *testing.T) {
	for _, keyword := range celKeywords {
		guard, expr := guardedPathCEL("o", model.Path{"spec", keyword}, true)
		for _, got := range []string{guard, expr} {
			if strings.Contains(got, "."+keyword) || strings.Contains(got, "has(o.spec."+keyword) {
				t.Errorf("the keyword %q must not be used as selector: %s", keyword, got)
			}
		}
	}
}

func Test_namePrefix(t *testing.T) {
	foo := "foo"
	tests := []struct {
		name       string
		namePrefix *string
		kind       string
		want       string
	}{
		{name: "derived from the kind", kind: "MachineDeployment", want: "machinedeployment_"},
		{name: "derived from a kind with an initialism", kind: "HTTPRoute", want: "httproute_"},
		{name: "the name prefix of the resource is ignored", namePrefix: &foo, kind: "MachineDeployment", want: "machinedeployment_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := model.Resource{NamePrefix: tt.namePrefix, GroupVersionKind: model.GroupVersionKind{Kind: tt.kind}}
			if got := namePrefix(resource); got != tt.want {
				t.Errorf("namePrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}
