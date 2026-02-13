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
package external_docs

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// ExternalDocSpec defines the desired state of ExternalDoc
type ExternalDocSpec struct {
	// This tests that external documentation can be attached to a field with url and description.
	// +kubebuilder:externalDoc:url="https://example.com/docs",description="external docs description"
	FieldWithExternalDoc string `json:"fieldWithExternalDoc,omitempty"`

	// This tests that external documentation can be attached with only url.
	// +kubebuilder:externalDoc:url="https://example.com/docs"
	FieldWithExternalDocURLOnly string `json:"fieldWithExternalDocURLOnly,omitempty"`

	// This tests that external documentation from a type is propagated.
	TypeWithExternalDoc TypeWithExternalDoc `json:"typeWithExternalDoc,omitempty"`
}

// TypeWithExternalDoc is a type with external documentation.
// +kubebuilder:externalDoc:url="https://example.com/type-docs",description="type-level external docs"
type TypeWithExternalDoc string

// ExternalDoc is the Schema for the external docs API
type ExternalDoc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ExternalDocSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ExternalDocList contains a list of ExternalDoc
type ExternalDocList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalDoc `json:"items"`
}
