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

package markers

import (
	"reflect"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

// The label marker gets applied to the custom resource, which is only done for markers
// on the type of the custom resource. Markers on fields would silently get ignored.
func Test_labelMarker_onlyRegisteredForTypes(t *testing.T) {
	found := false
	for _, def := range MarkerDefinitions {
		if def.Name != labelMarkerName {
			continue
		}
		found = true
		if def.Target != markers.DescribesType {
			t.Errorf("marker %s must only be registered for types, got target %v", labelMarkerName, def.Target)
		}
	}
	if !found {
		t.Errorf("marker %s is not registered", labelMarkerName)
	}
}

func Test_labelMarker_ApplyToResource(t *testing.T) {
	type fields struct {
		Name     string
		JSONPath jsonPath
	}
	tests := []struct {
		name         string
		fields       fields
		resource     *model.Resource
		wantResource *model.Resource
		wantErr      bool
	}{
		{
			name:         "happy path",
			fields:       fields{Name: "foo", JSONPath: ".bar"},
			resource:     &model.Resource{},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"bar"}}}},
		},
		{
			name:         "other label exists",
			fields:       fields{Name: "foo", JSONPath: ".bar"},
			resource:     &model.Resource{Labels: []model.Label{{Name: "other", Path: model.Path{"x"}}}},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "other", Path: model.Path{"x"}}, {Name: "foo", Path: model.Path{"bar"}}}},
		},
		{
			name:         "label already exists with identical path",
			fields:       fields{Name: "foo", JSONPath: ".bar"},
			resource:     &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"bar"}}}},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"bar"}}}},
		},
		{
			name:         "label already exists with same path length",
			fields:       fields{Name: "foo", JSONPath: ".bar"},
			resource:     &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"other"}}}},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"other"}}}},
			wantErr:      true,
		},
		{
			name:         "label already exists with different path length",
			fields:       fields{Name: "foo", JSONPath: ".bar"},
			resource:     &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"other", "path"}}}},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"other", "path"}}}},
			wantErr:      true,
		},
		{
			name:         "invalid json path",
			fields:       fields{Name: "foo", JSONPath: "{.bar}"},
			resource:     &model.Resource{},
			wantResource: &model.Resource{},
			wantErr:      true,
		},
		{
			name:     "nil resource",
			fields:   fields{Name: "foo", JSONPath: "{.bar}"},
			resource: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := labelMarker{
				Name:     tt.fields.Name,
				JSONPath: tt.fields.JSONPath,
			}
			if err := n.ApplyToResource(tt.resource); (err != nil) != tt.wantErr {
				t.Errorf("labelMarker.ApplyToResource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.resource, tt.wantResource) {
				t.Errorf("labelMarker.ApplyToResource() = %v, want %v", tt.resource, tt.wantResource)
			}
		})
	}
}
