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

	// The Go field is not named Help like the marker field, as this would collide with the Help method generated for the marker help.

	// MetricHelp specifies the help text for the metric.
	MetricHelp string `marker:"help,optional"`

	// Keys from the MetricMeta struct.

	// Labels specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.
	// Note: The label names group, version, kind, name and namespace can't be used, as the targets add labels with these names to every metric themselves. Using one of them is an error.
	// Note: If the path points to a map, labels for its entries are not supported, as they can't be aligned with the values by all targets.
	Labels map[string]jsonPath `marker:"labels,optional"`

	// Path specifies the JSONPath of the metric, relative to the field or type the marker is set on.
	// With the kube-state-metrics target the full path ends up in the path field of the metric.
	Path jsonPath `marker:"path,optional"`

	// Keys from the MetricStateSet struct.

	// States specifies a list of values to compare the given Value against.
	States []string `marker:"states"`

	// LabelName specifies the key of the label which is used for each entry in States to expose the value.
	// Note: It is required by the resource-state-metrics target.
	LabelName string `marker:"labelName,optional"`

	// IgnoreMissing specifies to leave the metric out silently if a field on its path doesn't exist, e.g. because the status is not set yet.
	// Note: Only used by the resource-state-metrics target. Without it the evaluation of the expression fails for such a resource, which resource-state-metrics logs and counts as an error.
	IgnoreMissing bool `marker:"ignoreMissing,optional"`

	// Value specifies the JSONPath to the field which gets used as value to compare against the states for equality, relative to the metric path.
	Value *jsonPath `marker:"value,optional"`
}

var _ LocalGeneratorMarker = &stateSetMarker{}

func (s stateSetMarker) ToGenerator(basePath ...string) (*model.Generator, error) {
	path, err := newPath(basePath, s.Path)
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
		Name:          s.Name,
		Help:          s.MetricHelp,
		Type:          model.MetricTypeStateSet,
		Path:          path,
		Labels:        labels,
		Value:         value,
		List:          s.States,
		LabelName:     s.LabelName,
		IgnoreMissing: s.IgnoreMissing,
	}, nil
}
