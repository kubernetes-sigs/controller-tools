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

package linters

import (
	"reflect"
	"sort"
	"testing"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/diff"
)

func TestMaxItemsArrays_Execute(t *testing.T) {
	crd := &v1.CustomResourceDefinition{
		Spec: v1.CustomResourceDefinitionSpec{
			PreserveUnknownFields: true,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Properties: map[string]v1.JSONSchemaProps{
								"test-string-field":            {Type: "string"},
								"test-field-without-max-items": {Type: "array"},
								"any-of-field": {AnyOf: []v1.JSONSchemaProps{
									{
										Type: "string",
									},
									{
										Type: "integer",
									},
									{
										Type: "array",
									},
								}},
							},
						},
					},
				},
			},
		},
	}

	expectedErrors := newWarningList(
		"spec.versions[0].schema.openAPIV3Schema.properties.any-of-field.anyOf[2].maxItems is not specified on a field of type 'array'",
		"spec.versions[0].schema.openAPIV3Schema.properties.test-field-without-max-items.maxItems is not specified on a field of type 'array'",
	)

	fieldEvaluator := MaxItemsArrays{}
	errs := fieldEvaluator.Execute(crd)
	// sort strings to allow for consistent comparison
	sort.Sort(expectedErrors)
	sort.Sort(errs)

	if !reflect.DeepEqual(errs, expectedErrors) {
		t.Errorf("returned errors were not as expected: %s", diff.ObjectGoPrintSideBySide(errs, expectedErrors))
	}
}
