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
package multiplefiles

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	multiver "testdata.kubebuilder.io/cronjob/multiple_versions"
)

type OneResourceSpec struct {
	Struct multiver.OuterStruct `json:"struct,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=oneresource

type OneResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec OneResourceSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type OneResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OneResource `json:"items"`
}
