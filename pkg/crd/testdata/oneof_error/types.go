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
// +versionName=v1beta1
package oneof_error

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=oneof

// OneofSpec is the spec for the oneofs API.
// +kubebuilder:validation:AtMostOneOf=field.foo;field.bar
type OneofSpec struct {
	Field *TypeWithOneofs `json:"field,omitempty"`
}

type TypeWithOneofs struct {
	Foo *string `json:"foo,omitempty"`
	Bar *string `json:"bar,omitempty"`
}

// Oneof is the Schema for the Oneof API
type Oneof struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec OneofSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// OneofList contains a list of Oneof
type OneofList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Oneof `json:"items"`
}
