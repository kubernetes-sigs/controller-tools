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
	resource := model.Resource{
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
	}

	want := Resource{
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
	}

	if diff := cmp.Diff(want, fromModel(resource)); diff != "" {
		t.Errorf("fromModel() (-want,+got): %s", diff)
	}
}

func Test_fromModel_unsetValues(t *testing.T) {
	got := fromModel(model.Resource{
		GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
		Metrics: []model.Generator{{
			Name: "i", Help: "info", Type: model.MetricTypeInfo,
			// A nil path must stay nil and an empty list of labels must result in an empty map, which is omitted.
			Path: nil, Labels: []model.Label{},
		}},
	})

	if got.MetricNamePrefix != nil {
		t.Errorf("expected the metric name prefix to stay nil, got %q", *got.MetricNamePrefix)
	}
	if len(got.Labels.LabelsFromPath) != 0 {
		t.Errorf("expected no resource labels, got %v", got.Labels.LabelsFromPath)
	}
	info := got.Metrics[0].Each.Info
	if info == nil || info.Path != nil || len(info.LabelsFromPath) != 0 {
		t.Errorf("expected nil path and no labels, got %#v", info)
	}
}
