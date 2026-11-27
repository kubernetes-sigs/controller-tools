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
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

// MarkerDefinitions are the markers of the state-metrics-config generator.
// The markers of the crd generator, which are needed to parse the types (e.g. groupName), are
// not part of it. They get registered by the generator from the definitions of the crd generator.
var MarkerDefinitions = []*markerDefinitionWithHelp{}

// ResourceMarker is a marker that configures a custom resource.
type ResourceMarker interface {
	// ApplyToResource applies this marker to the given resource.
	// It's called after the metrics of the resource are populated.
	ApplyToResource(resource *model.Resource) error
}

// LocalGeneratorMarker is a marker that defines a metric of a custom resource.
type LocalGeneratorMarker interface {
	// ToGenerator creates the metric. The basePath is the relative path to the field the marker is set on.
	ToGenerator(basePath ...string) (*model.Generator, error)
}
