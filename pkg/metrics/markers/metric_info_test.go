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

func Test_infoMarker_ToGenerator(t *testing.T) {
	tests := []struct {
		name       string
		infoMarker infoMarker
		basePath   []string
		want       *model.Generator
	}{
		{
			name:       "Happy path",
			infoMarker: infoMarker{},
			basePath:   []string{},
			want: &model.Generator{
				Type:   model.MetricTypeInfo,
				Path:   model.Path{},
				Labels: []model.Label{},
			},
		},
		{
			name: "keyLabel, relative JSONPath and labels",
			infoMarker: infoMarker{
				JSONPath: ".bar",
				KeyLabel: "key",
				Labels:   map[string]jsonPath{"b": ".b", "a": "."},
			},
			basePath: []string{"spec"},
			want: &model.Generator{
				Type: model.MetricTypeInfo,
				Path: model.Path{"spec", "bar"},
				// Labels are sorted by name.
				Labels:   []model.Label{{Name: "a", Path: model.Path{}}, {Name: "b", Path: model.Path{"b"}}},
				KeyLabel: "key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.infoMarker.ToGenerator(tt.basePath...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("infoMarker.ToGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}
