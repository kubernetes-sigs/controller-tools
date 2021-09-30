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

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/diff"
)

func TestMaxLengthString_Execute(t *testing.T) {
	oneVal := int64(1)
	crd := &v1.CustomResourceDefinition{
		Spec: v1.CustomResourceDefinitionSpec{
			PreserveUnknownFields: true,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Properties: map[string]v1.JSONSchemaProps{
								"test-field-with-max-length":    {Type: "string", MaxLength: &oneVal},
								"test-field-without-max-length": {Type: "string"},
								"any-of-field": {AnyOf: []v1.JSONSchemaProps{
									{
										Type: "string",
									},
									{
										Type: "integer",
									},
								}},
								"test-array-fields": {
									Type: "array",
									Items: &v1.JSONSchemaPropsOrArray{
										Schema: &v1.JSONSchemaProps{
											Properties: map[string]v1.JSONSchemaProps{
												"string-field": {Type: "string"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expectedErrors := []string{
		"spec.versions[0].schema.openAPIV3Schema.properties.any-of-field.anyOf[0].maxLength is not specified on a field of type 'string'",
		"spec.versions[0].schema.openAPIV3Schema.properties.test-array-fields.items.properties.string-field.maxLength is not specified on a field of type 'string'",
		"spec.versions[0].schema.openAPIV3Schema.properties.test-field-without-max-length.maxLength is not specified on a field of type 'string'",
	}

	fieldEvaluator := MaxLengthStrings{}
	errs := fieldEvaluator.Execute(crd)
	// sort strings to allow for consistent comparison
	sort.Strings(expectedErrors)
	sort.Strings(errs)

	if !reflect.DeepEqual(errs, expectedErrors) {
		t.Errorf("returned errors were not as expected: %s", diff.ObjectGoPrintSideBySide(errs, expectedErrors))
	}
}
