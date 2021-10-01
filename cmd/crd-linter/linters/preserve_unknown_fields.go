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

// PreserveUnknownFields checks if a CRD contains a 'preserveUnknownFields' marker
// either at the top level, or on any specific sub-field of the resource.
type PreserveUnknownFields struct{}

var _ Linter = PreserveUnknownFields{}

func (p PreserveUnknownFields) Name() string {
	return "PreserveUnknownFields"
}

func (p PreserveUnknownFields) Description() string {
	return "Fields should avoid using 'preserveUnknownFields' as it means any data can be persisted into the Kubernetes apiserver " +
		"without any guards on the size and type of data. Setting this to true is no longer permitted in the 'v1' API version of " +
		"CustomResourceDefinitions, and as such, it should not be used: https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#field-pruning"
}

func (p PreserveUnknownFields) Execute(crd *v1.CustomResourceDefinition) []string {
	var errs []string
	if crd.Spec.PreserveUnknownFields {
		errs = append(errs, "spec.preserveUnknownFields is set to 'true'")
	}
	return append(errs, recurseAllSchemas(crd.Spec.Versions, func(props v1.JSONSchemaProps, path string) []string {
		if props.XPreserveUnknownFields != nil && *props.XPreserveUnknownFields == true {
			return []string{fmt.Sprintf("%s.x-kubernetes-preserve-unknown-fields is set to 'true'", path)}
		}
		return nil
	})...)
}
