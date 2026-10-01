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

package metrics

import (
	"slices"
	"testing"

	ctrlmarkers "sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/config"
)

type fakeGeneratorMarker struct {
	name string
}

func (f fakeGeneratorMarker) ToGenerator(_ ...string) (*config.Generator, error) {
	return &config.Generator{Name: f.name}, nil
}

func Test_generatorsFromMarkers_deterministicOrder(t *testing.T) {
	markerValues := ctrlmarkers.MarkerValues{
		"k8s:controller-gen:metrics:stateset": {fakeGeneratorMarker{name: "stateset"}},
		"k8s:controller-gen:metrics:info":     {fakeGeneratorMarker{name: "info-1"}, fakeGeneratorMarker{name: "info-2"}},
		"k8s:controller-gen:metrics:gauge":    {fakeGeneratorMarker{name: "gauge"}},
	}
	want := []string{"gauge", "info-1", "info-2", "stateset"}

	// Iterating over a map is random, so run multiple times to make a flaky ordering visible.
	for range 50 {
		generators, err := generatorsFromMarkers(markerValues)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, 0, len(generators))
		for _, g := range generators {
			got = append(got, g.Name)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("generatorsFromMarkers() order = %v, want %v", got, want)
		}
	}
}

func Test_addPathPrefixOnGenerator(t *testing.T) {
	newGauge := func() config.Generator {
		return config.Generator{Each: config.Metric{
			Type:  config.MetricTypeGauge,
			Gauge: &config.MetricGauge{MetricMeta: config.MetricMeta{Path: []string{"bar"}}},
		}}
	}

	tests := []struct {
		name   string
		prefix []string
		want   []string
	}{
		{name: "prefix", prefix: []string{"foo"}, want: []string{"foo", "bar"}},
		{name: "no prefix for inlined fields", prefix: nil, want: []string{"bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addPathPrefixOnGenerator(newGauge(), tt.prefix).Each.Gauge.Path
			if !slices.Equal(got, tt.want) {
				t.Errorf("addPathPrefixOnGenerator() path = %v, want %v", got, tt.want)
			}
		})
	}
}
