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

func Test_gaugeMarker_ToGenerator(t *testing.T) {
	tests := []struct {
		name        string
		gaugeMarker gaugeMarker
		basePath    []string
		want        *model.Generator
	}{
		{
			name: "Happy path",
			gaugeMarker: gaugeMarker{
				Value: jsonPathPointer(".foo"),
			},
			basePath: []string{},
			want: &model.Generator{
				Type:   model.MetricTypeGauge,
				Path:   model.Path{},
				Labels: []model.Label{},
				Value:  model.Path{"foo"},
			},
		},
		{
			name: "keyLabel, missingAsZero and relative JSONPath",
			gaugeMarker: gaugeMarker{
				JSONPath:      ".bar",
				KeyLabel:      "key",
				MissingAsZero: true,
				Labels:        map[string]jsonPath{"l": ".l"},
			},
			basePath: []string{"spec"},
			want: &model.Generator{
				Type:          model.MetricTypeGauge,
				Path:          model.Path{"spec", "bar"},
				Labels:        []model.Label{{Name: "l", Path: model.Path{"l"}}},
				KeyLabel:      "key",
				MissingAsZero: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.gaugeMarker.ToGenerator(tt.basePath...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gaugeMarker.ToGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}
