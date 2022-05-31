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
package job

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testdata.kubebuilder.io/cronjob/unserved"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=job

// JobSpec is the spec for the jobs API.
type JobSpec struct {
	// FriendlyName is the friendly name for the job.
	//
	// +kubebuilder:validation:MinLength=5
	FriendlyName string `json:"friendlyName"`

	// Count is the number of times a job may be executed.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	Count int32 `json:"count"`

	// CronJob is the spec for the related CrongJob.
	CronnJob unserved.CronJobSpec `json:"crongJob"`
}

// Job is the Schema for the jobs API
type Job struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JobSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}
