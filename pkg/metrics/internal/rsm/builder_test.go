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

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func TestBuilder_AddResource(t *testing.T) {
	// newResource returns a resource with a single metric having the name foo_info and the help Foo.
	newResource := func(modify func(*model.Resource, *model.Generator)) model.Resource {
		metric := model.Generator{
			Name: "foo_info", Help: "Foo.", Type: model.MetricTypeInfo,
			Path: model.Path{"spec"}, PathKind: model.PathKindObject,
		}
		resource := model.Resource{GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"}}
		if modify != nil {
			modify(&resource, &metric)
		}
		resource.Metrics = []model.Generator{metric}
		return resource
	}
	label := func(name string) model.Label { return model.Label{Name: name, Path: model.Path{"spec", name}} }

	tests := []struct {
		name        string
		resource    model.Resource
		wantErrPart string
	}{
		{name: "valid", resource: newResource(nil)},
		{
			name:        "help is required",
			resource:    newResource(func(_ *model.Resource, m *model.Generator) { m.Help = "" }),
			wantErrPart: "requires the help to be set",
		},
		{
			name:        "name has to match the pattern",
			resource:    newResource(func(_ *model.Resource, m *model.Generator) { m.Name = "foo-info" }),
			wantErrPart: "requires the name to match",
		},
		{
			// The name prefix is part of the name of the family which has to match the pattern.
			name: "name prefix is part of the name",
			resource: newResource(func(r *model.Resource, _ *model.Generator) {
				prefix := "foo-bar"
				r.NamePrefix = &prefix
			}),
			wantErrPart: "foo-bar_foo_info",
		},
		{
			name: "labels with the name of an injected label are dropped",
			resource: newResource(func(r *model.Resource, m *model.Generator) {
				m.Labels = []model.Label{label("name"), label("kind")}
				r.Labels = []model.Label{label("namespace")}
			}),
		},
		{
			name: "key label of a map",
			resource: newResource(func(_ *model.Resource, m *model.Generator) {
				m.Path, m.PathKind, m.KeyLabel = model.Path{"spec", "m"}, model.PathKindMap, "key"
			}),
		},
		{
			name:        "key label of something not being a map",
			resource:    newResource(func(_ *model.Resource, m *model.Generator) { m.KeyLabel = "key" }),
			wantErrPart: "keyLabel can only be used if the path points to a map",
		},
		{
			name: "key label with the name of an injected label",
			resource: newResource(func(_ *model.Resource, m *model.Generator) {
				m.Path, m.PathKind, m.KeyLabel = model.Path{"spec", "m"}, model.PathKindMap, "name"
			}),
			wantErrPart: `the label "name" of the option keyLabel is added by resource-state-metrics`,
		},
		{
			name: "state label with the name of an injected label",
			resource: newResource(func(_ *model.Resource, m *model.Generator) {
				m.Type, m.LabelName, m.List = model.MetricTypeStateSet, "kind", []string{"a"}
			}),
			wantErrPart: `the label "kind" of the option labelName is added by resource-state-metrics`,
		},
		{
			name: "empty state label",
			resource: newResource(func(_ *model.Resource, m *model.Generator) {
				m.Type, m.List = model.MetricTypeStateSet, []string{"a"}
			}),
			wantErrPart: "the name of the label of the option labelName must not be empty",
		},
		{
			name: "duplicate label in a metric",
			resource: newResource(func(_ *model.Resource, m *model.Generator) {
				m.Labels = []model.Label{label("a"), label("a")}
			}),
			wantErrPart: `the label "a" is set multiple times`,
		},
		{
			name: "label of the metric and of the resource",
			resource: newResource(func(r *model.Resource, m *model.Generator) {
				m.Labels = []model.Label{label("a")}
				r.Labels = []model.Label{label("a")}
			}),
			wantErrPart: `the label "a" is set multiple times`,
		},
		{
			name: "state label and label",
			resource: newResource(func(_ *model.Resource, m *model.Generator) {
				m.Type, m.LabelName, m.List = model.MetricTypeStateSet, "a", []string{"a"}
				m.Labels = []model.Label{label("a")}
			}),
			wantErrPart: `the label "a" is set multiple times`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder("", "")
			err := b.AddResource(tt.resource)
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatal(err)
				}
				if len(b.stores) != 1 {
					t.Errorf("expected one store, got %d", len(b.stores))
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("AddResource() error = %v, want error containing %q", err, tt.wantErrPart)
			}
			if len(b.stores) != 0 {
				t.Errorf("expected no store to be added on error, got %d", len(b.stores))
			}
		})
	}
}
