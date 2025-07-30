/*
Copyright 2024 The Kubernetes Authors.

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
package enum

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:enum
type EnumType string

const (
	Value1 EnumType = "Value1"
	Value2 EnumType = "Value2"
)

// +kubebuilder:object:root=true

// Enum is a test CRD that contains an enum.
type Enum struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EnumSpec `json:"spec,omitempty"`
}

// EnumSpec defines the desired state of Enum
type EnumSpec struct {
	Field EnumType `json:"field,omitempty"`
}

// +kubebuilder:object:root=true

// EnumList contains a list of Enum
type EnumList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Enum `json:"items"`
}
