/*
Copyright The Kubernetes Authors.

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

// +groupName=testdata.kubebuilder.io
// +versionName=v1
package immutable

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type ImmutableType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ImmutableSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type ImmutableTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImmutableType `json:"items"`
}

type ImmutableSpec struct {
	// +k8s:immutable
	RequiredString string `json:"requiredString"`

	// +k8s:immutable
	// +optional
	OptionalString *string `json:"optionalString,omitempty"`

	// +k8s:immutable
	// +optional
	OptionalLiteralString string `json:"optionalLiteralString,omitempty"`

	// +k8s:immutable
	RequiredStruct NestedStruct `json:"requiredStruct"`

	// +k8s:immutable
	// +optional
	OptionalStruct *NestedStruct `json:"optionalStruct,omitempty"`

	// +k8s:immutable
	// +optional
	OptionalOmitZeroStruct NestedStruct `json:"optionalOmitZeroStruct,omitzero"`

	// +k8s:immutable
	RequiredSlice []string `json:"requiredSlice"`

	// +k8s:immutable
	// +optional
	OptionalSlice []string `json:"optionalSlice,omitempty"`

	// +k8s:immutable
	RequiredMap map[string]string `json:"requiredMap"`

	// +k8s:immutable
	// +optional
	OptionalMap map[string]string `json:"optionalMap,omitempty"`

	MutableString string `json:"mutableString"`
}

type NestedStruct struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
