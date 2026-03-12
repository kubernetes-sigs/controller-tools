/*

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
package iface

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InterfaceFieldSpec struct {
	Bar any `json:"bar,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:singular=interfacefield

// InterfaceField is the Schema for the InterfaceField API
type InterfaceField struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InterfaceFieldSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// InterfaceFieldList contains a list of InterfaceField
type InterfaceFieldList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InterfaceField `json:"items"`
}
