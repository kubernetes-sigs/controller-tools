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
package oneof

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=oneof

// OneofSpec is the spec for the oneofs API.
type OneofSpec struct {
	FirstTypeWithOneof  *TypeWithOneofs         `json:"firstTypeWithOneof,omitempty"`
	SecondTypeWithOneof *TypeWithMultipleOneofs `json:"secondTypeWithOneof,omitempty"`

	FirstTypeWithExactOneof  *TypeWithExactOneofs         `json:"firstTypeWithExactOneof,omitempty"`
	SecondTypeWithExactOneof *TypeWithMultipleExactOneofs `json:"secondTypeWithExactOneof,omitempty"`

	TypeWithMultipleAtLeastOneofs *TypeWithMultipleAtLeastOneofs `json:"typeWithMultipleAtLeastOneOf,omitempty"`

	TypeWithAllOneOf *TypeWithAllOneofs `json:"typeWithAllOneOf,omitempty"`

	FirstCustomTypeAlias CustomTypeAlias `json:"firstCustomTypeAlias,omitempty"`

	// This verifies if the custom type alias XValidation is not duplicated.
	SecondCustomTypeAlias CustomTypeAlias `json:"secondCustomTypeAlias,omitempty"`
}

// +kubebuilder:validation:XValidation:message="only one of foo|bar may be set",rule="!(has(self.foo) && has(self.bar))"
// +kubebuilder:validation:AtMostOneOf=foo;bar
type TypeWithOneofs struct {
	Foo *string `json:"foo,omitempty"`
	Bar *string `json:"bar,omitempty"`
}

// +kubebuilder:validation:AtMostOneOf=a;b
// +kubebuilder:validation:AtMostOneOf=c;d
type TypeWithMultipleOneofs struct {
	A *string `json:"a,omitempty"`
	B *string `json:"b,omitempty"`

	C *string `json:"c,omitempty"`
	D *string `json:"d,omitempty"`
}

// +kubebuilder:validation:ExactlyOneOf=foo;bar
type TypeWithExactOneofs struct {
	Foo *string `json:"foo,omitempty"`
	Bar *string `json:"bar,omitempty"`
}

// +kubebuilder:validation:ExactlyOneOf=a;b
// +kubebuilder:validation:ExactlyOneOf=c;d
type TypeWithMultipleExactOneofs struct {
	A *string `json:"a,omitempty"`
	B *string `json:"b,omitempty"`

	C *string `json:"c,omitempty"`
	D *string `json:"d,omitempty"`
}

// +kubebuilder:validation:AtLeastOneOf=a;b
// +kubebuilder:validation:AtLeastOneOf=c;d
type TypeWithMultipleAtLeastOneofs struct {
	A *string `json:"a,omitempty"`
	B *string `json:"b,omitempty"`

	C *string `json:"c,omitempty"`
	D *string `json:"d,omitempty"`
}

// +kubebuilder:validation:AtMostOneOf=a;b
// +kubebuilder:validation:ExactlyOneOf=c;d
// +kubebuilder:validation:AtLeastOneOf=e;f
type TypeWithAllOneofs struct {
	A *string `json:"a,omitempty"`
	B *string `json:"b,omitempty"`

	C *string `json:"c,omitempty"`
	D *string `json:"d,omitempty"`

	E *string `json:"e,omitempty"`
	F *string `json:"f,omitempty"`
}

// CustomTypeAlias is a custom alias
// +kubebuilder:validation:XValidation:rule="self >= 100 && self <= 1000",message="invalid CustomTypeAlias value"
type CustomTypeAlias *int32

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
