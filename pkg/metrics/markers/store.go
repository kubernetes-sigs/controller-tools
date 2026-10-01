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
	// StoreMarkerName is the marker for a store, which is a custom resource type
	// (identified by its group, version and kind) to generate metrics for.
	// Without a set StoreMarkerName the generator will not generate any configuration
	// for this custom resource.
	StoreMarkerName = "k8s:controller-gen:metrics:store"
)

func init() {
	MarkerDefinitions = append(
		MarkerDefinitions,
		must(markers.MakeDefinition(StoreMarkerName, markers.DescribesType, storeMarker{})).
			help(storeMarker{}.Help()),
	)
}

// +controllertools:marker:generateHelp:category=Metrics

// storeMarker enables the creation of a custom resource configuration entry and uses the given prefix for the metrics if configured.
type storeMarker struct {
	// NamePrefix specifies the prefix for all metrics of this resource.
	// Note: This field directly maps to the metricNamePrefix field in the resource's custom resource configuration.
	NamePrefix string `marker:"namePrefix,optional"`
}

var _ ResourceMarker = storeMarker{}

func (n storeMarker) ApplyToResource(resource *config.Resource) error {
	if n.NamePrefix != "" {
		resource.MetricNamePrefix = &n.NamePrefix
	}
	return nil
}
