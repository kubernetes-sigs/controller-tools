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

// +groupName=skipdescpkg.example.com
// +versionName=v1
// +kubebuilder:skip:description
package skipdescpkg

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PackageLevelSkip shows package-level skip in action.
//
// This description won't appear due to the package marker.
// +kubebuilder:object:root=true
type PackageLevelSkip struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PackageLevelSkipSpec `json:"spec,omitempty"`
}

// PackageLevelSkipSpec defines the spec.
// This description also won't appear.
type PackageLevelSkipSpec struct {
	// FieldOne has docs that won't appear
	// due to the package-level marker.
	FieldOne string `json:"fieldOne"`

	// FieldTwo also has docs that won't appear.
	FieldTwo int `json:"fieldTwo"`

	// ComplexField has a nested type.
	// All descriptions in this package are skipped.
	ComplexField NestedType `json:"complexField"`
}

// NestedType is a nested type.
// This description won't appear either.
type NestedType struct {
	// Value has docs that won't appear.
	Value string `json:"value"`
}

// PackageLevelSkipList contains a list of PackageLevelSkip.
// +kubebuilder:object:root=true
type PackageLevelSkipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PackageLevelSkip `json:"items"`
}
