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

//go:generate ../../../../.run-controller-gen.sh webhook paths=. output:dir=.

// +groupName=testdata.kubebuilder.io
// +versionName=v1
package cronjob

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobTwoSpec defines the desired state of CronJobTwo
type CronJobTwoSpec struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`
}

// CronJobTwoStatus defines the observed state of CronJobTwo
type CronJobTwoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=mycronjobtwo

// CronJobTwo is the Schema for the cronjobtwos API
type CronJobTwo struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobTwoSpec   `json:"spec,omitempty"`
	Status CronJobTwoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobTwoList contains a list of CronJobTwo
type CronJobTwoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJobTwo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJobTwo{}, &CronJobTwoList{})
}
