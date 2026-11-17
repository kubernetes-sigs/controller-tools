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

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_labelMarker_ApplyToResource(t *testing.T) {
	type fields struct {
		Name  string
		Value jsonPath
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
			fields:       fields{Name: "foo", Value: ".bar"},
			resource:     &model.Resource{},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"bar"}}}},
		},
		{
			name:         "other label exists",
			fields:       fields{Name: "foo", Value: ".bar"},
			resource:     &model.Resource{Labels: []model.Label{{Name: "other", Path: model.Path{"x"}}}},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "other", Path: model.Path{"x"}}, {Name: "foo", Path: model.Path{"bar"}}}},
		},
		{
			name:         "label already exists",
			fields:       fields{Name: "foo", Value: ".bar"},
			resource:     &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"bar"}}}},
			wantResource: &model.Resource{Labels: []model.Label{{Name: "foo", Path: model.Path{"bar"}}}},
			wantErr:      true,
		},
		{
			name:         "invalid json path",
			fields:       fields{Name: "foo", Value: "{.bar}"},
			resource:     &model.Resource{},
			wantResource: &model.Resource{},
			wantErr:      true,
		},
		{
			name:     "nil resource",
			fields:   fields{Name: "foo", Value: "{.bar}"},
			resource: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := labelMarker{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
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
