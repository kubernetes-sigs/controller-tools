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

func TestSchemaProvided_Execute_WithNoSchema(t *testing.T) {
	crd := &v1.CustomResourceDefinition{
		Spec: v1.CustomResourceDefinitionSpec{
			PreserveUnknownFields: true,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:   "v1alpha1",
					Schema: nil,
				},
				{
					Name:   "v1alpha2",
					Schema: &v1.CustomResourceValidation{OpenAPIV3Schema: nil},
				},
				{
					Name:   "v1alpha3",
					Schema: &v1.CustomResourceValidation{OpenAPIV3Schema: &v1.JSONSchemaProps{}},
				},
			},
		},
	}

	expectedErrors := []string{
		"spec.versions[0] (v1alpha1) does not provide a schema",
		"spec.versions[1] (v1alpha2) does not provide a schema",
	}

	fieldEvaluator := SchemaProvided{}
	errs := fieldEvaluator.Execute(crd)
	// sort strings to allow for consistent comparison
	sort.Strings(expectedErrors)
	sort.Strings(errs)

	if !reflect.DeepEqual(errs, expectedErrors) {
		t.Errorf("returned errors were not as expected: %s", diff.ObjectGoPrintSideBySide(errs, expectedErrors))
	}
}
