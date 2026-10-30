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

package model

import (
	"strings"
	"testing"
)

func TestResource_Validate_resourceLabels(t *testing.T) {
	gvk := GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Foo"}
	if err := (Resource{GroupVersionKind: gvk, Labels: []Label{{Name: "app", Path: Path{"x"}}}}).Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	err := Resource{GroupVersionKind: gvk, Labels: []Label{{Name: "name", Path: Path{"metadata", "name"}}}}.Validate()
	if err == nil || !strings.Contains(err.Error(), `the label "name" of the option label is added to every metric by the targets`) {
		t.Fatalf("Validate() error = %v, want an error for the label name", err)
	}
}

func TestResource_Validate(t *testing.T) {
	label := func(name string) []Label { return []Label{{Name: name, Path: Path{"x"}}} }

	tests := []struct {
		name        string
		generator   Generator
		wantErrPart string
	}{
		{name: "info on a map without labels", generator: Generator{Type: MetricTypeInfo, PathKind: PathKindMap}},
		{name: "info on a map with one label", generator: Generator{Type: MetricTypeInfo, PathKind: PathKindMap, Labels: label("a")}},
		{
			name:        "info on a map with two labels",
			generator:   Generator{Type: MetricTypeInfo, PathKind: PathKindMap, Labels: append(label("a"), label("b")...)},
			wantErrPart: "supports at most one label",
		},
		{
			name:        "labels with the name of an added label",
			generator:   Generator{Type: MetricTypeInfo, PathKind: PathKindScalar, Labels: label("name")},
			wantErrPart: `the label "name" of the option labels is added to every metric by the targets`,
		},
		{
			name:        "labels with the name of a label added by resource-state-metrics",
			generator:   Generator{Type: MetricTypeInfo, PathKind: PathKindScalar, Labels: label("namespace")},
			wantErrPart: `the label "namespace" of the option labels`,
		},
		{
			name:        "state label with the name of an added label",
			generator:   Generator{Type: MetricTypeStateSet, PathKind: PathKindScalar, LabelName: "kind"},
			wantErrPart: `the label "kind" of the option labelName`,
		},
		{name: "gauge on a map without labels", generator: Generator{Type: MetricTypeGauge, PathKind: PathKindMap}},
		{
			name:        "gauge on a map with a label",
			generator:   Generator{Type: MetricTypeGauge, PathKind: PathKindMap, Labels: label("a")},
			wantErrPart: "a Gauge metric on a map doesn't support labels",
		},
		{name: "stateset on a map without labels", generator: Generator{Type: MetricTypeStateSet, PathKind: PathKindMap}},
		{
			name:        "stateset on a map with a label",
			generator:   Generator{Type: MetricTypeStateSet, PathKind: PathKindMap, Labels: label("a")},
			wantErrPart: "a StateSet metric on a map doesn't support labels",
		},
		{
			name:      "lists and structs are not restricted",
			generator: Generator{Type: MetricTypeGauge, PathKind: PathKindArray, Labels: append(label("a"), label("b")...)},
		},
		{
			name:      "scalars are not restricted",
			generator: Generator{Type: MetricTypeInfo, PathKind: PathKindScalar, Labels: append(label("a"), label("b")...)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.generator.Name = "foo"
			err := Resource{
				GroupVersionKind: GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Foo"},
				Metrics:          []Generator{tt.generator},
			}.Validate()
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("Validate() error = %v, want error containing %q", err, tt.wantErrPart)
			}
		})
	}
}
