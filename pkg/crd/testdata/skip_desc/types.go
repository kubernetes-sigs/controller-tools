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

// +groupName=skipdesc.example.com
// +versionName=v1
package skipdesc

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SkipDescExample shows how skip description markers work.
//
// This type keeps its description since no skip marker was used.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type SkipDescExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkipDescExampleSpec   `json:"spec,omitempty"`
	Status SkipDescExampleStatus `json:"status,omitempty"`
}

// SkipDescExampleSpec defines the spec.
type SkipDescExampleSpec struct {
	// NormalField keeps its description.
	// This is a multi-line description to test that
	// descriptions work normally when no skip marker is present.
	NormalField string `json:"normalField"`

	// SkippedField won't have this description in the CRD.
	// +kubebuilder:skip:description
	SkippedField string `json:"skippedField"`

	// AnotherNormalField keeps its description.
	// This field is used to verify that the skip marker only affects
	// the specific field it's applied to.
	AnotherNormalField int `json:"anotherNormalField"`

	// ComplexField won't have this description.
	// +kubebuilder:skip:description
	ComplexField NestedType `json:"complexField"`
}

// NestedType is a nested structure.
// This description appears since it's at the type level.
type NestedType struct {
	// NestedField has docs that appear in the CRD.
	NestedField string `json:"nestedField"`

	// AnotherField also has docs.
	AnotherField int `json:"anotherField"`
}

// SkipDescExampleStatus defines the status.
type SkipDescExampleStatus struct {
	// Conditions represent the latest observations.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// SkipDescExampleList contains a list of SkipDescExample.
// +kubebuilder:object:root=true
type SkipDescExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkipDescExample `json:"items"`
}
