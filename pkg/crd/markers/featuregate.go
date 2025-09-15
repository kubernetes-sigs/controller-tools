/*
Copyright 2025.

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
)

// +controllertools:marker:generateHelp:category="CRD feature gates"

// FeatureGate marks a field to be conditionally included based on feature gate enablement.
// Fields marked with +kubebuilder:featuregate will only be included in generated CRDs
// when the specified feature gate is enabled via the crd:featureGates parameter.
type FeatureGate string

// ApplyToSchema does nothing for feature gates - they are processed by the generator
// to conditionally include/exclude fields.
func (FeatureGate) ApplyToSchema(schema *apiext.JSONSchemaProps, field string) error {
	// Feature gates don't modify the schema directly.
	// They are processed by the generator to conditionally include/exclude fields.
	return nil
}
