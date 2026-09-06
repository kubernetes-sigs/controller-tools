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
package declarativevalidation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type DeclarativeValidation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DeclarativeValidationSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type DeclarativeValidationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeclarativeValidation `json:"items"`
}

type DeclarativeValidationSpec struct {
	// +k8s:maxItems=2
	Items []string `json:"items"`

	// +k8s:minimum=-2
	// +k8s:maximum=2 # Bounds the CEL validation cost.
	Value int32 `json:"value"`
}
