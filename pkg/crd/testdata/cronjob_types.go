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

//go:generate controller-gen crd paths=. output:dir=.

// +groupName=testdata.kubebuilder.io
// +versionName=v1
package cronjob

import (
	"fmt"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobSpec defines the desired state of CronJob
type CronJobSpec struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// This tests that non-serialized fields aren't included in the schema.
	InternalData string `json:"-"`

	// This flag is like suspend, but for when you really mean it.
	// It helps test the +kubebuilder:validation:Type marker.
	// +optional
	NoReallySuspend *TotallyABool `json:"noReallySuspend,omitempty"`

	// This tests byte slice schema generation.
	BinaryName []byte `json:"binaryName"`

	// This tests that nullable works correctly
	// +nullable
	CanBeNull string `json:"canBeNull"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate"`

	// The number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// The number of failed finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`

	// This tests byte slices are allowed as map values.
	ByteSliceData map[string][]byte `json:"byteSliceData,omitempty"`

	// This tests string slices are allowed as map values.
	StringSliceData map[string][]string `json:"stringSliceData,omitempty"`

	// This tests that markers that are allowed on both fields and types are applied to fields
	// +kubebuilder:validation:MinLength=4
	TwoOfAKindPart0 string `json:"twoOfAKindPart0"`

	// This tests that markers that are allowed on both fields and types are applied to types
	TwoOfAKindPart1 LongerString `json:"twoOfAKindPart1"`
}

// +kubebuilder:validation:MinLength=4
// This tests that markers that are allowed on both fields and types are applied to types
type LongerString string

// use an explicit type marker to verify that apply-first markers generate properly

// +kubebuilder:validation:Type=string
// TotallyABool is a bool that serializes as a string.
type TotallyABool bool

func (t TotallyABool) MarshalJSON() ([]byte, error) {
	if t {
		return []byte(`"true"`), nil
	} else {
		return []byte(`"false"`), nil
	}
}
func (t *TotallyABool) UnmarshalJSON(in []byte) error {
	switch string(in) {
	case `"true"`:
		*t = true
		return nil
	case `"false"`:
		*t = false
	default:
		return fmt.Errorf("bad TotallyABool value %q", string(in))
	}
}

// ConcurrencyPolicy describes how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type ConcurrencyPolicy string

const (
	// AllowConcurrent allows CronJobs to run concurrently.
	AllowConcurrent ConcurrencyPolicy = "Allow"

	// ForbidConcurrent forbids concurrent runs, skipping next run if previous
	// hasn't finished yet.
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// CronJobStatus defines the observed state of CronJob
type CronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// A list of pointers to currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information about the last time the job was successfully scheduled,
	// with microsecond precision.
	// +optional
	LastScheduleMicroTime *metav1.MicroTime `json:"lastScheduleMicroTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSpec   `json:"spec,omitempty"`
	Status CronJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJob{}, &CronJobList{})
}
