/*
Copyright 2021 The Kubernetes Authors.

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

func init() {
	AllDefinitions = append(AllDefinitions,
		must(markers.MakeDefinition("kubebuilder:skipDescriptions", markers.DescribesField, SkipDescriptionsOnField{})).
			WithHelp(markers.SimpleHelp("CRD processing", "sets empty descriptions for nested properties of a field, recursively.")),

		must(markers.MakeDefinition("kubebuilder:skipDescriptions", markers.DescribesPackage, SkipDescriptionsOnPackage{})).
			WithHelp(markers.SimpleHelp("CRD processing", "sets empty descriptions for all types in a package, recursively.")),
	)
}

type SkipDescriptionsOnField struct{}

func (s SkipDescriptionsOnField) PostFlatten() {}

func (s SkipDescriptionsOnField) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	applyToSchemaProperties(schema)
	return nil
}

type SkipDescriptionsOnPackage struct{}

func (s SkipDescriptionsOnPackage) PostFlatten() {}

func (s SkipDescriptionsOnPackage) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	clearPropDescriptionRecurse(schema)
	return nil
}

func applyToSchemaProperties(schema *apiext.JSONSchemaProps) {
	for name := range schema.Properties {
		value := schema.Properties[name]
		clearPropDescriptionRecurse(&value)
		schema.Properties[name] = value
	}
}

func clearPropDescriptionRecurse(schema *apiext.JSONSchemaProps) {
	schema.Description = ""
	applyToSchemaProperties(schema)
}
