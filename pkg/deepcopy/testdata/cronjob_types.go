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

//go:generate controller-gen object paths=.

// +kubebuilder:object:generate=true
// +groupName=testdata.kubebuilder.io
// +versionName=v1
package cronjob

import (
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobSpec defines the desired state of CronJob
// Real-world test:
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

	// Specific cases
	SpecificCases SpecificCases `json:"specificCases"`
}

// SpecificCases excercises specific paths and/or regressions in the deepcopy machinery
type SpecificCases struct {
	// "Normal" Cases:

	// Reference type receiver: covered by SomePointers/StringMap
	// Non-reference-type receiver: covered by this struct

	// has manual deepcopy w/ pointer receiver
	ManualDeepCopyPtr DeepCopyPtr `json:"manualDeepCopyPtr"`
	// has manual deepcopy w/ non-pointer receiver
	ManualDeepCopyNonPtr DeepCopyNonPtr `json:"manualDeepCopyNonPtr"`

	// has manual deepcopyinto w/ pointer receiver
	ManualIntoPtr DeepCopyIntoPtr `json:"manualIntoPtr"`
	// has manual deepcopyinto w/ non-pointer receiver
	ManualIntoNonPtr DeepCopyIntoNonPtr `json:"manualIntoNonPtr"`

	// built-in types in fields
	BuiltInField string `json:"builtInField"`

	// non-quite deepcopies
	BadDeepCopyHasParams `json:"badDeepCopyHasParams"`
	BadDeepCopyNoReturn `json:"badDeepCopyNoReturn"`
	BadDeepCopyPtrVal `json:"badDeepCopyPtrVal"`
	BadDeepCopyNonPtrVal `json:"badDeepCopyNonPtrVal"`
	BadDeepCopyPtrMismatch `json:"badDeepCopyPtrMismatch"`

	// not-quite-deepcopyintos
	BadDeepCopyIntoNoParams `json:"badDeepCopyIntoNoParams"`
	BadDeepCopyIntoNonPtrParam `json:"badDeepCopyIntoNonPtrParam"`
	BadDeepCopyIntoHasResult `json:"badDeepCopyIntoHasResult"`

	// compound types in fields

	// also tests maps/slices/etc of built-in types
	MapInField map[string]string `json:"mapInField"`
	SliceInField []string `json:"sliceInField"`
	BuiltInPointer *string `json:"pointerInField"`
	NamedTypeInField TotallyAString `json:"namedTypeInField"`
	// named types to struct are tested via the reference to this object
	// named types to others are tested via the SomePointers/StringMap cases

	// other map types
	MapToDeepCopyPtr map[string]DeepCopyPtr `json:"mapToDeepCopyPtr"`
	MapToDeepCopyNonPtr map[string]DeepCopyPtr `json:"mapToDeepCopyNonPtr"`
	MapToDeepCopyIntoPtr map[string]DeepCopyIntoPtr `json:"mapToDeepCopyIntoPtr"`
	MapToDeepCopyIntoNonPtr map[string]DeepCopyIntoNonPtr `json:"mapToDeepCopyIntoNonPtr"`
	MapToShallowNamedType map[string]TotallyAString `json:"mapToShallowNamedType"`
	MapToReferenceType map[string][]string `json:"mapToReferenceType"`
	MapToStruct map[string]CronJobSpec `json:"mapToStruct"`
	MapWithNamedKeys map[TotallyAString]int `json:"mapWithNamedKeys"`
	MapToPtrToDeepCopyIntoRefType map[string]*DeepCopyIntoRef `json:"mapToPtrToDeepCopyIntoRefType"`
	MapToDeepCopyIntoRefType map[string]DeepCopyIntoRef `json:"mapToDeepCopyIntoRefType"`

	// other slice types
	SliceToDeepCopyPtr []DeepCopyPtr `json:"sliceToDeepCopyPtr"`
	SliceToDeepCopyNonPtr []DeepCopyPtr `json:"sliceToDeepCopyNonPtr"`
	SliceToDeepCopyIntoPtr []DeepCopyIntoPtr `json:"sliceToDeepCopyIntoPtr"`
	SliceToDeepCopyIntoNonPtr []DeepCopyIntoNonPtr `json:"sliceToDeepCopyIntoNonPtr"`
	SliceToShallowNamedType []TotallyAString `json:"sliceToShallowNamedType"`
	SliceToReferenceType [][]string `json:"sliceToReferenceType"`
	SliceToStruct []CronJobSpec `json:"sliceToStruct"`

	// other pointer types
	PtrToDeepCopyPtr *DeepCopyPtr `json:"ptrToDeepCopyPtr"`
	PtrToDeepCopyNonPtr *DeepCopyPtr `json:"ptrToDeepCopyNonPtr"`
	PtrToDeepCopyIntoPtr *DeepCopyIntoPtr `json:"ptrToDeepCopyIntoPtr"`
	PtrToDeepCopyIntoNonPtr *DeepCopyIntoNonPtr `json:"ptrToDeepCopyIntoNonPtr"`
	PtrToShallowNamedType *TotallyAString `json:"ptrToShallowNamedType"`
	PtrToReferenceType *[]string `json:"ptrToReferenceType"`
	PtrToStruct *CronJobSpec `json:"ptrToStruct"`
	PtrToDeepCopyIntoRef *DeepCopyIntoRef `json:"ptrToDeepCopyIntoRef"`


	// Regression Tests:

	// Case: kubernetes-sigs/controller-tools#262 part 1 (see type definition)
	SomePointers SliceOfPointers `json:"somePointers"`

	// Case: kubernetes-sigs/controller-tools#262 part 2 (see type definition)
	StringMap MapOfStrings `json:"stringMap"`
}

// Test aliases to basic types
type TotallyAString string

// Tests manual DeepCopy with a pointer receiver
type DeepCopyPtr struct {}
func (d *DeepCopyPtr) DeepCopy() *DeepCopyPtr {
	return &DeepCopyPtr{}
}

// Tests manual DeepCopy with a non-pointer receiver
type DeepCopyNonPtr struct {}
func (d DeepCopyNonPtr) DeepCopy() DeepCopyNonPtr {
	return DeepCopyNonPtr{}
}

// Tests manual DeepCopyInto with a pointer receiver
type DeepCopyIntoPtr struct {}
func (d *DeepCopyIntoPtr) DeepCopyInto(out *DeepCopyIntoPtr) {
	*out = DeepCopyIntoPtr{}
}

// Tests manual DeepCopyInto with a non-pointer receiver
type DeepCopyIntoNonPtr struct {}
func (d DeepCopyIntoNonPtr) DeepCopyInto(out *DeepCopyIntoNonPtr) {
	*out = DeepCopyIntoNonPtr{}
}

// Tests manual DeepCopyInto with a reference type receiver
type DeepCopyIntoRef map[string]string
func (d DeepCopyIntoRef) DeepCopyInto(out *DeepCopyIntoRef) {
	*out = make(DeepCopyIntoRef)
}

// Case: kubernetes-sigs/controller-tools#262 part 1:
// Type renames of slices to pointers.
type SliceOfPointers []*SomeStruct
type SomeStruct struct {
	Foo string `json:"foo"`
}

// Case: kubernetes-sigs/controller-tools#262 part 2:
// Type renames on "allowable" maps
type MapOfStrings map[string]string

// tests bad deep copy methods

type BadDeepCopyHasParams struct{}
func (d *BadDeepCopyHasParams) DeepCopy(out string) *BadDeepCopyHasParams {
	return nil
} 

type BadDeepCopyNoReturn struct{}
func (d *BadDeepCopyNoReturn) DeepCopy() {}

type BadDeepCopyPtrVal struct {}
func (d *BadDeepCopyPtrVal) DeepCopy() *BadDeepCopyNoReturn {
	return nil
}

type BadDeepCopyNonPtrVal struct {}
func (d BadDeepCopyNonPtrVal) DeepCopy() BadDeepCopyNoReturn {
	return BadDeepCopyNoReturn{}
}

type BadDeepCopyPtrMismatch struct {}
func (d BadDeepCopyPtrMismatch) DeepCopy() *BadDeepCopyPtrMismatch {
	return nil
}

type BadDeepCopyIntoNoParams struct {}
func (d BadDeepCopyIntoNoParams) DeepCopyInto() {}

type BadDeepCopyIntoNonPtrParam struct {}
func (d BadDeepCopyIntoNonPtrParam) DeepCopyInto(out BadDeepCopyIntoNonPtrParam) {}

type BadDeepCopyIntoHasResult struct {}
func (d BadDeepCopyIntoHasResult) DeepCopyInto(out *BadDeepCopyIntoHasResult) error {
	return nil
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
