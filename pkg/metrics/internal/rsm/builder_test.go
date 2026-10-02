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
	newResource := func(name, help string) model.Resource {
		return model.Resource{
			GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
			Metrics: []model.Generator{{
				Name: name, Help: help, Type: model.MetricTypeInfo,
				Path: model.Path{"spec"}, PathKind: model.PathKindObject,
			}},
		}
	}

	tests := []struct {
		name        string
		resource    model.Resource
		wantErrPart string
	}{
		{name: "valid", resource: newResource("foo_info", "Foo.")},
		{name: "help is required", resource: newResource("foo_info", ""), wantErrPart: "requires the help to be set"},
		{name: "name has to match the pattern", resource: newResource("foo-info", "Foo."), wantErrPart: "requires the name to match"},
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

func TestBuilder_AddResource_namePrefixIsPartOfTheName(t *testing.T) {
	prefix := "foo-bar"
	b := NewBuilder("", "")
	err := b.AddResource(model.Resource{
		NamePrefix: &prefix,
		Metrics:    []model.Generator{{Name: "info", Help: "Info.", Type: model.MetricTypeInfo}},
	})
	if err == nil || !strings.Contains(err.Error(), "foo-bar_info") {
		t.Fatalf("expected an error about the name including the prefix, got %v", err)
	}
}

func TestNewBuilder(t *testing.T) {
	if got := NewBuilder("", ""); got.name != defaultName || got.namespace != "" {
		t.Errorf("NewBuilder() = %q/%q, want the default name and no namespace", got.namespace, got.name)
	}
	if got := NewBuilder("foo", "bar"); got.name != "foo" || got.namespace != "bar" {
		t.Errorf("NewBuilder() = %q/%q, want bar/foo", got.namespace, got.name)
	}
}

func TestBuilder_AddResource_labels(t *testing.T) {
	newResource := func(metric model.Generator, labels ...model.Label) model.Resource {
		metric.Name, metric.Help = "foo", "Foo."
		return model.Resource{
			GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
			Labels:           labels,
			Metrics:          []model.Generator{metric},
		}
	}
	label := func(name string) model.Label { return model.Label{Name: name, Path: model.Path{"spec", name}} }

	tests := []struct {
		name        string
		resource    model.Resource
		wantErrPart string
	}{
		{
			name:     "labels with the name of an injected label are dropped",
			resource: newResource(model.Generator{Type: model.MetricTypeInfo, Labels: []model.Label{label("name"), label("kind")}}, label("namespace")),
		},
		{
			name:     "key label of a map",
			resource: newResource(model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"spec", "m"}, PathKind: model.PathKindMap, KeyLabel: "key"}),
		},
		{
			name:        "key label of something not being a map",
			resource:    newResource(model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"spec"}, PathKind: model.PathKindObject, KeyLabel: "key"}),
			wantErrPart: "keyLabel can only be used if the path points to a map",
		},
		{
			name:        "key label with the name of an injected label",
			resource:    newResource(model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"spec", "m"}, PathKind: model.PathKindMap, KeyLabel: "name"}),
			wantErrPart: `the label "name" of the option keyLabel is added by resource-state-metrics`,
		},
		{
			name:        "state label with the name of an injected label",
			resource:    newResource(model.Generator{Type: model.MetricTypeStateSet, LabelName: "kind", List: []string{"a"}}),
			wantErrPart: `the label "kind" of the option labelName is added by resource-state-metrics`,
		},
		{
			name:        "empty state label",
			resource:    newResource(model.Generator{Type: model.MetricTypeStateSet, List: []string{"a"}}),
			wantErrPart: "the name of the label of the option labelName must not be empty",
		},
		{
			name:        "duplicate label in a metric",
			resource:    newResource(model.Generator{Type: model.MetricTypeInfo, Labels: []model.Label{label("a"), label("a")}}),
			wantErrPart: `the label "a" is set multiple times`,
		},
		{
			name:        "label of the metric and of the resource",
			resource:    newResource(model.Generator{Type: model.MetricTypeInfo, Labels: []model.Label{label("a")}}, label("a")),
			wantErrPart: `the label "a" is set multiple times`,
		},
		{
			name:        "state label and label",
			resource:    newResource(model.Generator{Type: model.MetricTypeStateSet, LabelName: "a", List: []string{"a"}, Labels: []model.Label{label("a")}}),
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
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("AddResource() error = %v, want error containing %q", err, tt.wantErrPart)
			}
		})
	}
}

// The iteration order of the entries of a map differs between evaluations, so labels
// of the entries can't be aligned with the value.
