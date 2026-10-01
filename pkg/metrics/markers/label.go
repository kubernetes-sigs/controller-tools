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
	"errors"
	"fmt"

	"sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/config"
)

const (
	labelMarkerName = "k8s:controller-gen:metrics:label"
)

func init() {
	MarkerDefinitions = append(
		MarkerDefinitions,
		must(markers.MakeDefinition(labelMarkerName, markers.DescribesType, labelMarker{})).
			help(labelMarker{}.Help()),
	)
}

// +controllertools:marker:generateHelp:category=Metrics

// labelMarker specifies additional labels for all metrics of the custom resource.
// It is only considered on the type of the custom resource, which is the type having the store marker.
// The JSONPath is relative to the custom resource.
type labelMarker struct {
	// Name specifies the name of the label.
	Name string
	// JSONPath specifies the path to the value for the label, relative to the custom resource.
	JSONPath jsonPath `marker:"JSONPath"`
}

var _ ResourceMarker = labelMarker{}

func (n labelMarker) ApplyToResource(resource *config.Resource) error {
	if resource == nil {
		return errors.New("expected resource to not be nil")
	}

	jsonPathElems, err := n.JSONPath.Parse()
	if err != nil {
		return err
	}

	if resource.LabelsFromPath == nil {
		resource.LabelsFromPath = map[string][]string{}
	}

	if jsonPath, labelExists := resource.LabelsFromPath[n.Name]; labelExists {
		if len(jsonPathElems) != len(jsonPath) {
			return fmt.Errorf("duplicate definition for label %q", n.Name)
		}
		for i, v := range jsonPath {
			if v != jsonPathElems[i] {
				return fmt.Errorf("duplicate definition for label %q", n.Name)
			}
		}
	}

	resource.LabelsFromPath[n.Name] = jsonPathElems
	return nil
}
