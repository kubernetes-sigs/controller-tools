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

func TestNoPreserveUnknownFields_Execute_WithNoPreserveUnknownFields(t *testing.T) {
	trueVal := true
	falseVal := false
	crd := &v1.CustomResourceDefinition{
		Spec: v1.CustomResourceDefinitionSpec{
			PreserveUnknownFields: true,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Properties: map[string]v1.JSONSchemaProps{
								"test-field-with-preserve-fields":       {XPreserveUnknownFields: &trueVal},
								"test-field-with-false-preserve-fields": {XPreserveUnknownFields: &falseVal},
								"test-field-with-nil-preserve-fields":   {XPreserveUnknownFields: nil},
								"test-field-with-nested-preserve-fields": {
									Properties: map[string]v1.JSONSchemaProps{
										"nested-field": {XPreserveUnknownFields: &trueVal},
									},
								},
								"array-type-field": {
									Type: "array",
									Items: &v1.JSONSchemaPropsOrArray{
										Schema: &v1.JSONSchemaProps{
											XPreserveUnknownFields: &trueVal,
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

	expectedErrors := newWarningList(
		"spec.preserveUnknownFields is set to 'true'",
		"spec.versions[0].schema.openAPIV3Schema.properties.array-type-field.items.x-kubernetes-preserve-unknown-fields is set to 'true'",
		"spec.versions[0].schema.openAPIV3Schema.properties.test-field-with-nested-preserve-fields.properties.nested-field.x-kubernetes-preserve-unknown-fields is set to 'true'",
		"spec.versions[0].schema.openAPIV3Schema.properties.test-field-with-preserve-fields.x-kubernetes-preserve-unknown-fields is set to 'true'",
	)

	fieldEvaluator := NoPreserveUnknownFields{}
	errs := fieldEvaluator.Execute(crd)
	// sort strings to allow for consistent comparison
	sort.Sort(expectedErrors)
	sort.Sort(errs)

	if !reflect.DeepEqual(errs, expectedErrors) {
		t.Errorf("returned errors were not as expected: %s", diff.ObjectGoPrintSideBySide(errs, expectedErrors))
	}
}

func TestNoPreserveUnknownFields_Execute_WithNoNoPreserveUnknownFields(t *testing.T) {
	falseVal := false
	crd := &v1.CustomResourceDefinition{
		Spec: v1.CustomResourceDefinitionSpec{
			PreserveUnknownFields: false,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Properties: map[string]v1.JSONSchemaProps{
								"test-field-with-false-preserve-fields": {XPreserveUnknownFields: &falseVal},
							},
						},
					},
				},
			},
		},
	}

	fieldEvaluator := NoPreserveUnknownFields{}
	errs := fieldEvaluator.Execute(crd)

	if errs != nil {
		t.Errorf("returned errors were not as expected: %s", diff.ObjectGoPrintSideBySide(errs, nil))
	}
}
