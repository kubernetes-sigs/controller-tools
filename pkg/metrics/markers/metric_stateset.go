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
	stateSetMarkerName = "Metrics:stateset"
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

	// LabelsFromPath specifies additional labels where the value is taken from the given JSONPath.
	LabelsFromPath map[string]jsonPath `marker:"labelsFromPath,optional"`

	// Keys from the MetricStateSet struct.

	// List specifies a list of values to compare the given JSONPath against.
	List []string `marker:"list"`
	// LabelName specifies the key of the label which is used for each entry in List to expose the value.
	LabelName string `marker:"labelName,optional"`
	// JSONPath specifies the path to the field which gets used as value to compare against the list for equality.
	// Note: This field directly maps to the valueFrom field in the custom resource configuration.
	JSONPath *jsonPath `marker:"JSONPath,optional"`
}

var _ LocalGeneratorMarker = &stateSetMarker{}

func (s stateSetMarker) ToGenerator(basePath ...string) (*config.Generator, error) {
	var valueFrom []string
	var err error
	if s.JSONPath != nil {
		valueFrom, err = s.JSONPath.Parse()
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSONPath: %v", err)
		}
	}

	meta, err := newMetricMeta(basePath, "", s.LabelsFromPath)
	if err != nil {
		return nil, err
	}

	return &config.Generator{
		Name: s.Name,
		Help: s.MetricHelp,
		Each: config.Metric{
			Type: config.MetricTypeStateSet,
			StateSet: &config.MetricStateSet{
				MetricMeta: meta,
				List:       s.List,
				LabelName:  s.LabelName,
				ValueFrom:  valueFrom,
			},
		},
	}, nil
}
