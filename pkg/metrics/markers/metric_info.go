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
	"sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

const (
	infoMarkerName = "k8s:controller-gen:metrics:info"
)

func init() {
	MarkerDefinitions = append(
		MarkerDefinitions,
		must(markers.MakeDefinition(infoMarkerName, markers.DescribesField, infoMarker{})).
			help(infoMarker{}.Help()),
		must(markers.MakeDefinition(infoMarkerName, markers.DescribesType, infoMarker{})).
			help(infoMarker{}.Help()),
	)
}

// +controllertools:marker:generateHelp:category=Metric type Info

// infoMarker defines a Info metric and uses the implicit path to the field as path for the metric configuration.
// Info is a metric which is used to expose textual information.
// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#info
type infoMarker struct {
	// Keys from the Generator struct.

	// Name specifies the Name of the metric.
	Name string

	// The Go field is not named Help like the marker field, as this would collide with the Help method generated for the marker help.

	// MetricHelp specifies the help text for the metric.
	MetricHelp string `marker:"help,optional"`

	// Keys from the MetricMeta struct.

	// Labels specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.
	// Note: The label names group, version, kind, name and namespace can't be used, as the targets add labels with these names to every metric themselves. Using one of them is an error.
	// Note: If the path points to a map, at most one label for its entries is supported, as the labels can't be aligned with each other by all targets.
	Labels map[string]jsonPath `marker:"labels,optional"`

	// Path specifies the JSONPath of the metric, relative to the field or type the marker is set on.
	// With the kube-state-metrics target the full path ends up in the path field of the metric.
	Path jsonPath `marker:"path,optional"`

	// IgnoreMissing specifies to leave the metric out silently if a field on its path doesn't exist, e.g. because the status is not set yet.
	// Note: Only used by the resource-state-metrics target. Without it the evaluation of the expression fails for such a resource, which resource-state-metrics logs and counts as an error.
	IgnoreMissing bool `marker:"ignoreMissing,optional"`
}

var _ LocalGeneratorMarker = &infoMarker{}

func (i infoMarker) ToGenerator(basePath ...string) (*model.Generator, error) {
	path, err := newPath(basePath, i.Path)
	if err != nil {
		return nil, err
	}

	labels, err := newLabels(i.Labels)
	if err != nil {
		return nil, err
	}

	return &model.Generator{
		Name:          i.Name,
		Help:          i.MetricHelp,
		Type:          model.MetricTypeInfo,
		Path:          path,
		Labels:        labels,
		IgnoreMissing: i.IgnoreMissing,
	}, nil
}
