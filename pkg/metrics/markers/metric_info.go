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

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/config"
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
	// MetricHelp specifies the help text for the metric.
	MetricHelp string `marker:"help,optional"`

	// Keys from the MetricMeta struct.

	// Labels specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.
	Labels map[string]jsonPath `marker:"labels,optional"`
	// JSONPath specifies the relative path from this marker.
	// Note: This field get's appended to the path field in the custom resource configuration.
	JSONPath jsonPath `marker:"JSONPath,optional"`

	// Keys from the MetricInfo struct.

	// KeyLabel specifies a label which will be added to the metric having the object's key as value.
	// Note: This is only meaningful if the metric path points to an object (map). It is not validated.
	KeyLabel string `marker:"keyLabel,optional"`
}

var _ LocalGeneratorMarker = &infoMarker{}

func (i infoMarker) ToGenerator(basePath ...string) (*config.Generator, error) {
	meta, err := newMetricMeta(basePath, i.JSONPath, i.Labels)
	if err != nil {
		return nil, err
	}

	return &config.Generator{
		Name: i.Name,
		Help: i.MetricHelp,
		Each: config.Metric{
			Type: config.MetricTypeInfo,
			Info: &config.MetricInfo{
				MetricMeta:   meta,
				LabelFromKey: i.KeyLabel,
			},
		},
	}, nil
}
