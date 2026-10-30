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
	// Note: This field is only used by the kube-state-metrics target, where it directly maps to the metricNamePrefix field and replaces the default prefix kube_customresource, e.g. foo_created.
	// It is ignored by the resource-state-metrics target, which always uses the lowercased kind as prefix and adds it after the prefix kube_customresource_, e.g. kube_customresource_machinedeployment_created for the kind MachineDeployment.
	NamePrefix string `marker:"namePrefix,optional"`
}

var _ ResourceMarker = storeMarker{}

func (n storeMarker) ApplyToResource(resource *model.Resource) error {
	if n.NamePrefix != "" {
		resource.NamePrefix = &n.NamePrefix
	}
	return nil
}
