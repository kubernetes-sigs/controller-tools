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

package ksm

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_fromModel(t *testing.T) {
	prefix := "foo"

	tests := []struct {
		name     string
		resource model.Resource
		want     Resource
	}{
		{
			name: "all fields are converted",
			resource: model.Resource{
				NamePrefix:       &prefix,
				GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
				Labels:           []model.Label{{Name: "name", Path: model.Path{"metadata", "name"}}},
				Metrics: []model.Generator{
					{
						Name: "g", Help: "gauge", Type: model.MetricTypeGauge,
						Path: model.Path{"spec", "a"}, Value: model.Path{"b"}, KeyLabel: "key", MissingAsZero: true,
						Labels: []model.Label{{Name: "l", Path: model.Path{"l"}}},
					},
					{
						Name: "i", Help: "info", Type: model.MetricTypeInfo,
						Path: model.Path{"spec", "i"}, KeyLabel: "key",
					},
					{
						Name: "s", Help: "stateset", Type: model.MetricTypeStateSet,
						Path: model.Path{"spec", "s"}, Value: model.Path{"state"}, List: []string{"a", "b"}, LabelName: "state",
					},
				},
			},
			want: Resource{
				MetricNamePrefix: &prefix,
				GroupVersionKind: GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
				Labels:           Labels{LabelsFromPath: map[string][]string{"name": {"metadata", "name"}}},
				Metrics: []Generator{
					{Name: "g", Help: "gauge", Each: Metric{Type: MetricTypeGauge, Gauge: &MetricGauge{
						MetricMeta:   MetricMeta{LabelsFromPath: map[string][]string{"l": {"l"}}, Path: []string{"spec", "a"}},
						ValueFrom:    []string{"b"},
						LabelFromKey: "key",
						NilIsZero:    true,
					}}},
					{Name: "i", Help: "info", Each: Metric{Type: MetricTypeInfo, Info: &MetricInfo{
						MetricMeta:   MetricMeta{LabelsFromPath: map[string][]string{}, Path: []string{"spec", "i"}},
						LabelFromKey: "key",
					}}},
					{Name: "s", Help: "stateset", Each: Metric{Type: MetricTypeStateSet, StateSet: &MetricStateSet{
						MetricMeta: MetricMeta{LabelsFromPath: map[string][]string{}, Path: []string{"spec", "s"}},
						List:       []string{"a", "b"},
						LabelName:  "state",
						ValueFrom:  []string{"state"},
					}}},
				},
			},
		},
		{
			// A nil name prefix and path must stay nil and empty labels must result in an empty map.
			// Both are omitted in the output.
			name: "unset values",
			resource: model.Resource{
				GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
				Metrics: []model.Generator{{
					Name: "i", Help: "info", Type: model.MetricTypeInfo,
					Path: nil, Labels: []model.Label{},
				}},
			},
			want: Resource{
				GroupVersionKind: GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
				Labels:           Labels{LabelsFromPath: map[string][]string{}},
				Metrics: []Generator{{Name: "i", Help: "info", Each: Metric{Type: MetricTypeInfo, Info: &MetricInfo{
					MetricMeta: MetricMeta{LabelsFromPath: map[string][]string{}},
				}}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, fromModel(tt.resource)); diff != "" {
				t.Errorf("fromModel() (-want,+got): %s", diff)
			}
		})
	}
}
