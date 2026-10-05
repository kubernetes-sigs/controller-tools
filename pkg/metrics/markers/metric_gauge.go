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
	"fmt"

	"sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

const (
	gaugeMarkerName = "k8s:controller-gen:metrics:gauge"
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
// Arrays and objects will generate a metric per element and require Value to be set.
// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#gauge
type gaugeMarker struct {
	// Keys from the Generator struct.

	// Name specifies the Name of the metric.
	Name string
	// MetricHelp specifies the help text for the metric.
	MetricHelp string `marker:"help,optional"`

	// Keys from the MetricMeta struct.

	// Labels specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.
	// Note: With the resource-state-metrics target the label names group, version, kind, name and namespace are dropped, as resource-state-metrics adds them to every metric itself.
	// Note: If the path points to a map, labels for its entries are not supported, as they can't be aligned with the values by all targets.
	Labels map[string]jsonPath `marker:"labels,optional"`
	// JSONPath specifies the relative path from this marker.
	// Note: This field get's appended to the path field in the custom resource configuration.
	JSONPath jsonPath `marker:"JSONPath,optional"`

	// Keys from the MetricGauge struct.

	// Value specifies the JSONPath to a numeric field that will be the metric value, relative to the metric path.
	// Note: With the resource-state-metrics target, the value is converted to a unix timestamp if the field is a metav1.Time or metav1.MicroTime.
	Value *jsonPath `marker:"value,optional"`
	// KeyLabel specifies a label which will be added to the metric having the object's key as value.
	// Note: This is only meaningful if the metric path points to a map. With the resource-state-metrics target this is validated.
	// Note: If the path points to a map, labels for its entries, including this label, are not supported by gauge metrics, as they can't be aligned with the values by all targets.
	KeyLabel string `marker:"keyLabel,optional"`
	// MissingAsZero specifies to expose a not-existing field as zero value instead of omitting the metric.
	// Note: With the resource-state-metrics target this also applies to the entries of lists and maps having no value, which are not left out then.
	MissingAsZero bool `marker:"missingAsZero,optional"`
}

var _ LocalGeneratorMarker = &gaugeMarker{}

func (g gaugeMarker) ToGenerator(basePath ...string) (*model.Generator, error) {
	path, err := newPath(basePath, g.JSONPath)
	if err != nil {
		return nil, err
	}

	labels, err := newLabels(g.Labels)
	if err != nil {
		return nil, err
	}

	var value model.Path
	if g.Value != nil {
		value, err = g.Value.Parse()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %w", err)
		}
	}

	return &model.Generator{
		Name:          g.Name,
		Help:          g.MetricHelp,
		Type:          model.MetricTypeGauge,
		Path:          path,
		Labels:        labels,
		KeyLabel:      g.KeyLabel,
		Value:         value,
		MissingAsZero: g.MissingAsZero,
	}, nil
}
