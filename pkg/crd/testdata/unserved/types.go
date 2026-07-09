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
// +versionName=v1beta1
package unserved

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobSpec defines the desired state of CronJob
type CronJobSpec struct {
	// This tests that markers that are allowed on both fields and types are applied to fields
	// +kubebuilder:validation:MinLength=4
	TwoOfAKindPart0 string `json:"twoOfAKindPart0"`

	// +kubebuilder:validation:Minimum=-2
	// +kubebuilder:validation:Maximum=2
	// +kubebuilder:validation:MultipleOf=2
	Int32WithValidations int32 `json:"int32WithValidations"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=mycronjob
// +kubebuilder:unservedversion

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CronJobSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}
