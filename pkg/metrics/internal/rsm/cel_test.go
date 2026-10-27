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
		NamePrefix:       &prefix,
		GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
		Labels: []model.Label{
			{Name: "cluster", Path: model.Path{"metadata", "labels", "cluster"}},
			// The label name is injected by resource-state-metrics, so it is dropped automatically.
			{Name: "name", Path: model.Path{"metadata", "name"}},
		},
		Metrics: []model.Generator{
			{
				Name: "condition", Help: "A condition.", Type: model.MetricTypeStateSet,
				Path: model.Path{"status", "conditions"}, PathKind: model.PathKindArray,
				Labels:    []model.Label{{Name: "type", Path: model.Path{"type"}}},
				List:      []string{"True", "False"},
				LabelName: "status",
				Value:     model.Path{"status"},
			},
			{
				Name: "transition", Help: "Transition time.", Type: model.MetricTypeGauge,
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
				Name: "foo_condition", Help: "A condition.",
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
				Name: "foo_transition", Help: "Transition time.",
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

func Test_stateSetToCEL(t *testing.T) {
	tests := []struct {
		name   string
		metric model.Generator
		want   []Metric
	}{
		{
			name: "scalar",
			metric: model.Generator{
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
			metric: model.Generator{
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
			metric: model.Generator{
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
			metric: model.Generator{
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
			metric: model.Generator{
				Path: model.Path{"metadata", "ownerReferences"}, PathKind: model.PathKindArray,
				Labels: []model.Label{{Name: "owner_kind", Path: model.Path{"kind"}}},
			},
			want: Metric{
				Labels: []Label{{Name: "owner_kind", Value: `has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, has(v.kind) ? string(v.kind) : "") : []`}},
				Value:  `has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, 1) : []`,
			},
		},
		{
			name:   "map with a key label",
			metric: model.Generator{Path: model.Path{"spec", "labels"}, PathKind: model.PathKindMap, KeyLabel: "key"},
			want: Metric{
				Labels: []Label{{Name: "key", Value: `has(o.spec) && has(o.spec.labels) ? o.spec.labels.map(k, k) : []`}},
				Value:  `has(o.spec) && has(o.spec.labels) ? o.spec.labels.map(k, 1) : []`,
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
			metric: model.Generator{Path: model.Path{"spec", "replicas"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.spec) && has(o.spec.replicas) ? [o.spec.replicas] : []`},
		},
		{
			name:   "missing as zero",
			metric: model.Generator{Path: model.Path{"spec", "paused"}, PathKind: model.PathKindScalar, MissingAsZero: true},
			want:   Metric{Value: `has(o.spec) && has(o.spec.paused) ? o.spec.paused : 0`},
		},
		{
			name:   "value path of an object",
			metric: model.Generator{Path: model.Path{"status", "ref"}, Value: model.Path{"count"}, PathKind: model.PathKindObject},
			want:   Metric{Value: `has(o.status) && has(o.status.ref) && has(o.status.ref.count) ? [o.status.ref.count] : []`},
		},
		{
			name:   "list elements without the value are left out",
			metric: model.Generator{Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray},
			want:   Metric{Value: `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, v.size) : []`},
		},
		{
			name:   "list elements without the value are zero with missing as zero",
			metric: model.Generator{Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray, MissingAsZero: true},
			want:   Metric{Value: `has(o.status) && has(o.status.items) ? o.status.items.map(v, has(v.size) ? v.size : 0) : []`},
		},
		{
			name:   "list of values",
			metric: model.Generator{Path: model.Path{"status", "sizes"}, PathKind: model.PathKindArray},
			want:   Metric{Value: `has(o.status) && has(o.status.sizes) ? o.status.sizes.map(v, v) : []`},
		},
		{
			// The labels use the same filter as the value, as the entries of the lists have to be aligned.
			name: "list labels use the filter of the value",
			metric: model.Generator{
				Path: model.Path{"status", "items"}, PathKind: model.PathKindArray, Value: model.Path{"size"},
				// The label name is injected by resource-state-metrics, so it is dropped automatically.
				Labels: []model.Label{{Name: "name", Path: model.Path{"name"}}, {Name: "id", Path: model.Path{"id"}}},
			},
			want: Metric{
				Labels: []Label{{Name: "id", Value: `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, has(v.id) ? string(v.id) : "") : []`}},
				Value:  `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, v.size) : []`,
			},
		},
		{
			name:   "map entries without the value are left out",
			metric: model.Generator{Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap},
			want:   Metric{Value: `has(o.status) && has(o.status.pools) ? o.status.pools.filter(k, has(o.status.pools[k].size)).map(k, o.status.pools[k].size) : []`},
		},
		{
			name:   "map entries without the value are zero with missing as zero",
			metric: model.Generator{Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap, MissingAsZero: true},
			want:   Metric{Value: `has(o.status) && has(o.status.pools) ? o.status.pools.map(k, has(o.status.pools[k].size) ? o.status.pools[k].size : 0) : []`},
		},
		{
			name:   "map of values",
			metric: model.Generator{Path: model.Path{"status", "counts"}, PathKind: model.PathKindMap},
			want:   Metric{Value: `has(o.status) && has(o.status.counts) ? o.status.counts.map(k, o.status.counts[k]) : []`},
		},
		{
			// The key label has to use the same elements as the value.
			name:   "map with a key label",
			metric: model.Generator{Path: model.Path{"spec", "pools"}, PathKind: model.PathKindMap, Value: model.Path{"size"}, KeyLabel: "pool"},
			want: Metric{
				Labels: []Label{{Name: "pool", Value: `has(o.spec) && has(o.spec.pools) ? o.spec.pools.filter(k, has(o.spec.pools[k].size)).map(k, k) : []`}},
				Value:  `has(o.spec) && has(o.spec.pools) ? o.spec.pools.filter(k, has(o.spec.pools[k].size)).map(k, o.spec.pools[k].size) : []`,
			},
		},
		{
			name:   "timestamp type in a map",
			metric: model.Generator{Path: model.Path{"status", "at"}, Value: model.Path{"time"}, PathKind: model.PathKindMap, ValueKind: model.ValueKindTimestamp},
			want:   Metric{Value: `has(o.status) && has(o.status.at) ? o.status.at.filter(k, has(o.status.at[k].time)).map(k, unixSeconds(o.status.at[k].time)) : []`},
		},
		{
			name:   "timestamp type",
			metric: model.Generator{Path: model.Path{"status", "at"}, PathKind: model.PathKindScalar, ValueKind: model.ValueKindTimestamp},
			want:   Metric{Value: `has(o.status) && has(o.status.at) ? [unixSeconds(o.status.at)] : []`},
		},
		{
			// The conversion would fail for numeric fields like these.
			name:   "no timestamp conversion based on the field name",
			metric: model.Generator{Path: model.Path{"status", "uptime"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.status) && has(o.status.uptime) ? [o.status.uptime] : []`},
		},
		{
			name:   "field names which are CEL keywords",
			metric: model.Generator{Path: model.Path{"spec", "in", "null"}, PathKind: model.PathKindScalar},
			want:   Metric{Value: `has(o.spec) && "in" in o.spec && "null" in o.spec["in"] ? [o.spec["in"]["null"]] : []`},
		},
		{
			name:   "non identifier field name",
			metric: model.Generator{Path: model.Path{"metadata", "labels", "cluster.x-k8s.io/name"}, PathKind: model.PathKindScalar},
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
		guard, expr := guardedPathCEL("o", model.Path{"spec", keyword})
		for _, got := range []string{guard, expr} {
			if strings.Contains(got, "."+keyword) || strings.Contains(got, "has(o.spec."+keyword) {
				t.Errorf("the keyword %q must not be used as selector: %s", keyword, got)
			}
		}
	}
}

func Test_labelsToCEL_dropsInjectedLabels(t *testing.T) {
	got := labelsToCEL([]model.Label{
		{Name: "group"}, {Name: "version"}, {Name: "kind"}, {Name: "name"}, {Name: "namespace"},
		{Name: "owner", Path: model.Path{"spec", "owner"}},
	}, nil, model.PathKindScalar, "")
	if len(got) != 1 || got[0].Name != "owner" {
		t.Errorf("expected only the label owner, got %#v", got)
	}
}
