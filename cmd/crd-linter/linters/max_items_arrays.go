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

// MaxItemsArrays checks all array type fields on a CRD to ensure they have
// a 'maxItems' specified.
type MaxItemsArrays struct{}

func (m MaxItemsArrays) Name() string {
	return "MaxItemsArrays"
}

func (m MaxItemsArrays) Execute(crd *v1.CustomResourceDefinition) WarningList {
	return recurseAllSchemas(crd.Spec.Versions, func(props v1.JSONSchemaProps, path string) []Warning {
		if props.Type == "array" && props.MaxItems == nil {
			return newWarningList(fmt.Sprintf("%s.maxItems is not specified on a field of type 'array'", path))
		}
		return nil
	})
}

func (m MaxItemsArrays) Description() string {
	return "All 'array' typed fields should have a 'maxItems' specified, even if arbitrarily high, to ensure you do " +
		"not accidentally store more data that originally intended. This allows the apiserver to be sure you do not " +
		"store too much data in the apiserver, which can potentially lead to instability."
}

var _ Linter = MaxItemsArrays{}
