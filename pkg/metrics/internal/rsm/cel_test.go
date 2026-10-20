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

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_toStore(t *testing.T) {
	prefix := "foo"
	resource := model.Resource{
		NamePrefix:       &prefix,
		GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
		Labels: []model.Label{
			{Name: "cluster", Path: model.Path{"metadata", "labels", "cluster"}},
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

	got := toStore(resource)
	if got.Resource != "policies" || got.Resolver != resolverCEL {
		t.Fatalf("unexpected store: %#v", got)
	}
	if len(got.Labels) != 0 {
		t.Errorf("expected the labels of the resource to be on the metrics, got %#v on the store", got.Labels)
	}
	if got.Families[0].Name != "foo_condition" {
		t.Errorf("expected the name prefix to be added to the family name, got %q", got.Families[0].Name)
	}

	labels := got.Families[0].Metrics[0].Labels
	// type, status and the label of the resource. The label `name` is injected by resource-state-metrics.
	if len(labels) != 3 {
		t.Fatalf("expected the automatically injected name label to be omitted, got %#v", labels)
	}
	if diff := cmp.Diff(`has(o.metadata) && has(o.metadata.labels) && has(o.metadata.labels.cluster) ? string(o.metadata.labels.cluster) : ""`, labels[2].Value); diff != "" {
		t.Errorf("resource label expression (-want,+got): %s", diff)
	}
	if diff := cmp.Diff(`has(o.status) && has(o.status.conditions) ? o.status.conditions.filter(v, has(v.status)).map(v, int(v.status == "True" ? 1 : 0)) : []`, got.Families[0].Metrics[0].Value); diff != "" {
		t.Errorf("state set value expression (-want,+got): %s", diff)
	}
	if diff := cmp.Diff(`has(o.status) && has(o.status.conditions) ? o.status.conditions.filter(v, has(v.lastTransitionTime)).map(v, unixSeconds(v.lastTransitionTime)) : []`, got.Families[1].Metrics[0].Value); diff != "" {
		t.Errorf("timestamp value expression (-want,+got): %s", diff)
	}
}

func Test_stateSetToCEL_scalar(t *testing.T) {
	got := stateSetToCEL(model.Generator{
		Path: model.Path{"status", "phase"}, PathKind: model.PathKindScalar,
		List: []string{"Ready"}, LabelName: "phase",
	})
	if diff := cmp.Diff(`has(o.status) && has(o.status.phase) ? [int(o.status.phase == "Ready" ? 1 : 0)] : []`, got[0].Value); diff != "" {
		t.Errorf("scalar state set expression (-want,+got): %s", diff)
	}
}

func Test_infoToCEL(t *testing.T) {
	got := infoToCEL(model.Generator{
		Path:     model.Path{"status", "nodeRef"},
		PathKind: model.PathKindObject,
		Labels:   []model.Label{{Name: "node_name", Path: model.Path{"name"}}},
	})
	if diff := cmp.Diff(`has(o.status) && has(o.status.nodeRef) && has(o.status.nodeRef.name) ? string(o.status.nodeRef.name) : ""`, got.Labels[0].Value); diff != "" {
		t.Errorf("object label expression (-want,+got): %s", diff)
	}
	if diff := cmp.Diff(`has(o.status) && has(o.status.nodeRef) ? [1] : []`, got.Value); diff != "" {
		t.Errorf("object value expression (-want,+got): %s", diff)
	}

	list := infoToCEL(model.Generator{
		Path:     model.Path{"metadata", "ownerReferences"},
		PathKind: model.PathKindArray,
		Labels:   []model.Label{{Name: "owner_kind", Path: model.Path{"kind"}}},
	})
	if diff := cmp.Diff(`has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, 1) : []`, list.Value); diff != "" {
		t.Errorf("list value expression (-want,+got): %s", diff)
	}
	if diff := cmp.Diff(`has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, has(v.kind) ? string(v.kind) : "") : []`, list.Labels[0].Value); diff != "" {
		t.Errorf("list label expression (-want,+got): %s", diff)
	}
}

func Test_gaugeToCEL(t *testing.T) {
	tests := []struct {
		name   string
		metric model.Generator
		want   string
	}{
		{
			name:   "scalar",
			metric: model.Generator{Path: model.Path{"spec", "replicas"}, PathKind: model.PathKindScalar},
			want:   `has(o.spec) && has(o.spec.replicas) ? [o.spec.replicas] : []`,
		},
		{
			name:   "missing as zero",
			metric: model.Generator{Path: model.Path{"spec", "paused"}, PathKind: model.PathKindScalar, MissingAsZero: true},
			want:   `has(o.spec) && has(o.spec.paused) ? o.spec.paused : 0`,
		},
		{
			name:   "value path of an object",
			metric: model.Generator{Path: model.Path{"status", "ref"}, Value: model.Path{"count"}, PathKind: model.PathKindObject},
			want:   `has(o.status) && has(o.status.ref) && has(o.status.ref.count) ? [o.status.ref.count] : []`,
		},
		{
			name:   "list elements without the value are left out",
			metric: model.Generator{Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray},
			want:   `has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, v.size) : []`,
		},
		{
			name:   "list elements without the value are zero with missing as zero",
			metric: model.Generator{Path: model.Path{"status", "items"}, Value: model.Path{"size"}, PathKind: model.PathKindArray, MissingAsZero: true},
			want:   `has(o.status) && has(o.status.items) ? o.status.items.map(v, has(v.size) ? v.size : 0) : []`,
		},
		{
			name:   "list of values",
			metric: model.Generator{Path: model.Path{"status", "sizes"}, PathKind: model.PathKindArray},
			want:   `has(o.status) && has(o.status.sizes) ? o.status.sizes.map(v, v) : []`,
		},
		{
			name:   "map entries without the value are left out",
			metric: model.Generator{Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap},
			want:   `has(o.status) && has(o.status.pools) ? o.status.pools.filter(k, has(o.status.pools[k].size)).map(k, o.status.pools[k].size) : []`,
		},
		{
			name:   "map entries without the value are zero with missing as zero",
			metric: model.Generator{Path: model.Path{"status", "pools"}, Value: model.Path{"size"}, PathKind: model.PathKindMap, MissingAsZero: true},
			want:   `has(o.status) && has(o.status.pools) ? o.status.pools.map(k, has(o.status.pools[k].size) ? o.status.pools[k].size : 0) : []`,
		},
		{
			name:   "map of values",
			metric: model.Generator{Path: model.Path{"status", "counts"}, PathKind: model.PathKindMap},
			want:   `has(o.status) && has(o.status.counts) ? o.status.counts.map(k, o.status.counts[k]) : []`,
		},
		{
			name:   "timestamp type in a map",
			metric: model.Generator{Path: model.Path{"status", "at"}, Value: model.Path{"time"}, PathKind: model.PathKindMap, ValueKind: model.ValueKindTimestamp},
			want:   `has(o.status) && has(o.status.at) ? o.status.at.filter(k, has(o.status.at[k].time)).map(k, unixSeconds(o.status.at[k].time)) : []`,
		},
		{
			name:   "timestamp type",
			metric: model.Generator{Path: model.Path{"status", "at"}, PathKind: model.PathKindScalar, ValueKind: model.ValueKindTimestamp},
			want:   `has(o.status) && has(o.status.at) ? [unixSeconds(o.status.at)] : []`,
		},
		{
			// The conversion would fail for numeric fields like these.
			name:   "no timestamp conversion based on the field name",
			metric: model.Generator{Path: model.Path{"status", "uptime"}, PathKind: model.PathKindScalar},
			want:   `has(o.status) && has(o.status.uptime) ? [o.status.uptime] : []`,
		},
		{
			name:   "field names which are CEL keywords",
			metric: model.Generator{Path: model.Path{"spec", "in", "null"}, PathKind: model.PathKindScalar},
			want:   `has(o.spec) && "in" in o.spec && "null" in o.spec["in"] ? [o.spec["in"]["null"]] : []`,
		},
		{
			name:   "non identifier field name",
			metric: model.Generator{Path: model.Path{"metadata", "labels", "cluster.x-k8s.io/name"}, PathKind: model.PathKindScalar},
			want:   `has(o.metadata) && has(o.metadata.labels) && "cluster.x-k8s.io/name" in o.metadata.labels ? [o.metadata.labels["cluster.x-k8s.io/name"]] : []`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, gaugeToCEL(tt.metric).Value); diff != "" {
				t.Errorf("gaugeToCEL() value (-want,+got): %s", diff)
			}
		})
	}
}

func Test_keyLabelToCEL(t *testing.T) {
	info := infoToCEL(model.Generator{Path: model.Path{"spec", "labels"}, PathKind: model.PathKindMap, KeyLabel: "key"})
	if len(info.Labels) != 1 || info.Labels[0].Name != "key" {
		t.Fatalf("expected the key label, got %#v", info.Labels)
	}
	if diff := cmp.Diff(`has(o.spec) && has(o.spec.labels) ? o.spec.labels.map(k, k) : []`, info.Labels[0].Value); diff != "" {
		t.Errorf("info key label expression (-want,+got): %s", diff)
	}
	if diff := cmp.Diff(`has(o.spec) && has(o.spec.labels) ? o.spec.labels.map(k, 1) : []`, info.Value); diff != "" {
		t.Errorf("info value expression (-want,+got): %s", diff)
	}

	// The key label of a gauge has to use the same elements as the value.
	gauge := gaugeToCEL(model.Generator{Path: model.Path{"spec", "pools"}, PathKind: model.PathKindMap, Value: model.Path{"size"}, KeyLabel: "pool"})
	const items = `has(o.spec) && has(o.spec.pools) ? o.spec.pools.filter(k, has(o.spec.pools[k].size))`
	if diff := cmp.Diff(items+`.map(k, k) : []`, gauge.Labels[0].Value); diff != "" {
		t.Errorf("gauge key label expression (-want,+got): %s", diff)
	}
	if diff := cmp.Diff(items+`.map(k, o.spec.pools[k].size) : []`, gauge.Value); diff != "" {
		t.Errorf("gauge value expression (-want,+got): %s", diff)
	}
}

func Test_stateSetToCEL_map(t *testing.T) {
	got := stateSetToCEL(model.Generator{
		Path: model.Path{"status", "pools"}, PathKind: model.PathKindMap,
		Value: model.Path{"phase"}, List: []string{"Ready"}, LabelName: "phase",
	})
	const items = `has(o.status) && has(o.status.pools) ? o.status.pools.filter(k, has(o.status.pools[k].phase))`
	if diff := cmp.Diff(items+`.map(k, int(o.status.pools[k].phase == "Ready" ? 1 : 0)) : []`, got[0].Value); diff != "" {
		t.Errorf("map state set value expression (-want,+got): %s", diff)
	}
}

func Test_listLabelsUseTheFilterOfTheValue(t *testing.T) {
	got := gaugeToCEL(model.Generator{
		Path: model.Path{"status", "items"}, PathKind: model.PathKindArray, Value: model.Path{"size"},
		Labels: []model.Label{{Name: "name", Path: model.Path{"name"}}, {Name: "id", Path: model.Path{"id"}}},
	})
	// The label name is injected by resource-state-metrics.
	if len(got.Labels) != 1 {
		t.Fatalf("expected one label, got %#v", got.Labels)
	}
	if diff := cmp.Diff(`has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, has(v.id) ? string(v.id) : "") : []`, got.Labels[0].Value); diff != "" {
		t.Errorf("label expression (-want,+got): %s", diff)
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
