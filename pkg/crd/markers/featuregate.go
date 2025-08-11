/*
Copyright 2025 The Kubernetes Authors.

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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// +controllertools:marker:generateHelp:category="CRD feature gates"
// FeatureGate marks a field or type to be conditionally included based on feature gate enablement.
// The field will only be included in generated CRDs when the specified feature gate is enabled.
type FeatureGate string

// ApplyToSchema implements SchemaMarker interface.
// This marker doesn't directly modify the schema - it's used by the generator
// to conditionally include/exclude fields during CRD generation.
func (m FeatureGate) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	// Feature gate markers don't modify the schema directly.
	// They are processed by the generator to conditionally include/exclude fields.
	return nil
}

// Help returns the help information for this marker.
func (FeatureGate) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "CRD feature gates",
		DetailedHelp: markers.DetailedHelp{
			Summary: "marks a field to be conditionally included based on feature gate enablement",
			Details: "Fields marked with +kubebuilder:feature-gate will only be included in generated CRDs when the specified feature gate is enabled via --feature-gates flag.",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"": {
				Summary: "the name of the feature gate that controls this field",
				Details: "The feature gate name should match gates passed via --feature-gates=gate1=true,gate2=false",
			},
		},
	}
}
