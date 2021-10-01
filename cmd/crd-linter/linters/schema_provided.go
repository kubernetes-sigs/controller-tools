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
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// SchemaProvided verifies that all API versions provide a schema
type SchemaProvided struct{}

var _ Linter = SchemaProvided{}

func (p SchemaProvided) Name() string {
	return "SchemaProvided"
}

func (p SchemaProvided) Description() string {
	return "Not providing a schema is no longer possible with v1 CRDs, and means many API server features such as " +
		"server-side apply are not possible. Additionally, custom resources contents cannot be validated and arbitrary " +
		"data can be stored in objects, leading to potential API server instability."
}

func (p SchemaProvided) Execute(crd *v1.CustomResourceDefinition) []string {
	var errs []string
	for i, vers := range crd.Spec.Versions {
		if vers.Schema == nil || vers.Schema.OpenAPIV3Schema == nil {
			errs = append(errs, fmt.Sprintf("spec.versions[%d] (%s) does not provide a schema", i, vers.Name))
		}
	}
	return errs
}
