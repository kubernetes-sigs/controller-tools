/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"fmt"

	"sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/config"
)

const (
	gaugeMarkerName = "Metrics:gauge"
)

func init() {
	MarkerDefinitions = append(
		MarkerDefinitions,
		must(markers.MakeDefinition(gaugeMarkerName, markers.DescribesField, gaugeMarker{})).
			help(gaugeMarker{}.Help()),
		must(markers.MakeDefinition(gaugeMarkerName, markers.DescribesType, gaugeMarker{})).
			help(gaugeMarker{}.Help()),
	)
}

// +controllertools:marker:generateHelp:category=Metric type Gauge

// gaugeMarker defines a Gauge metric and uses the implicit path to the field joined by the provided JSONPath as path for the metric configuration.
// Gauge is a metric which targets a Path that may be a single value, array, or object.
// Arrays and objects will generate a metric per element and requre ValueFrom to be set.
// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#gauge
type gaugeMarker struct {
	// Keys from the Generator struct.

	// Name specifies the Name of the metric.
	Name string
	// MetricHelp specifies the help text for the metric.
	MetricHelp string `marker:"help,optional"`

	// Keys from the MetricMeta struct.

	// LabelsFromPath specifies additional labels where the value is taken from the given JSONPath.
	LabelsFromPath map[string]jsonPath `marker:"labelsFromPath,optional"`
	// JSONPath specifies the relative path from this marker.
	// Note: This field get's appended to the path field in the custom resource configuration.
	JSONPath jsonPath `marker:"JSONPath,optional"`

	// Keys from the MetricGauge struct.

	// ValueFrom specifies the JSONPath to a numeric field that will be the metric value.
	ValueFrom *jsonPath `marker:"valueFrom,optional"`
	// LabelFromKey specifies a label which will be added to the metric having the object's key as value.
	LabelFromKey string `marker:"labelFromKey,optional"`
	// NilIsZero specifies to treat a not-existing field as zero value.
	NilIsZero bool `marker:"nilIsZero,optional"`
}

var _ LocalGeneratorMarker = &gaugeMarker{}

func (g gaugeMarker) ToGenerator(basePath ...string) (*config.Generator, error) {
	var err error
	var valueFrom []string
	if g.ValueFrom != nil {
		valueFrom, err = g.ValueFrom.Parse()
		if err != nil {
			return nil, fmt.Errorf("failed to parse ValueFrom: %w", err)
		}
	}

	meta, err := newMetricMeta(basePath, g.JSONPath, g.LabelsFromPath)
	if err != nil {
		return nil, err
	}

	return &config.Generator{
		Name: g.Name,
		Help: g.MetricHelp,
		Each: config.Metric{
			Type: config.MetricTypeGauge,
			Gauge: &config.MetricGauge{
				NilIsZero:    g.NilIsZero,
				MetricMeta:   meta,
				LabelFromKey: g.LabelFromKey,
				ValueFrom:    valueFrom,
			},
		},
	}, nil
}
