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
	stateSetMarkerName = "k8s:controller-gen:metrics:stateset"
)

func init() {
	MarkerDefinitions = append(
		MarkerDefinitions,
		must(markers.MakeDefinition(stateSetMarkerName, markers.DescribesField, stateSetMarker{})).
			help(stateSetMarker{}.Help()),
		must(markers.MakeDefinition(stateSetMarkerName, markers.DescribesType, stateSetMarker{})).
			help(stateSetMarker{}.Help()),
	)
}

// +controllertools:marker:generateHelp:category=Metric type StateSet

// stateSetMarker defines a StateSet metric and uses the implicit path to the field as path for the metric configuration.
// A StateSet is a metric which represent a series of related boolean values, also called a bitset.
// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#stateset
type stateSetMarker struct {
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

	// Keys from the MetricStateSet struct.

	// List specifies a list of values to compare the given Value against.
	List []string `marker:"list"`
	// LabelName specifies the key of the label which is used for each entry in List to expose the value.
	LabelName string `marker:"labelName,optional"`
	// Value specifies the JSONPath to the field which gets used as value to compare against the list for equality, relative to the metric path.
	Value *jsonPath `marker:"value,optional"`
}

var _ LocalGeneratorMarker = &stateSetMarker{}

func (s stateSetMarker) ToGenerator(basePath ...string) (*model.Generator, error) {
	path, err := newPath(basePath, s.JSONPath)
	if err != nil {
		return nil, err
	}

	labels, err := newLabels(s.Labels)
	if err != nil {
		return nil, err
	}

	var value model.Path
	if s.Value != nil {
		value, err = s.Value.Parse()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %w", err)
		}
	}

	return &model.Generator{
		Name:      s.Name,
		Help:      s.MetricHelp,
		Type:      model.MetricTypeStateSet,
		Path:      path,
		Labels:    labels,
		Value:     value,
		List:      s.List,
		LabelName: s.LabelName,
	}, nil
}
