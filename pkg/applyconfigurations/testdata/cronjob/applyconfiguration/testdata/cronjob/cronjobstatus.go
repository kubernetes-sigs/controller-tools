// Code generated by controller-gen. DO NOT EDIT.

package cronjob

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobStatusApplyConfiguration represents an declarative configuration of the CronJobStatus type for use
// with apply.
type CronJobStatusApplyConfiguration struct {
	Active                []v1.ObjectReference `json:"active,omitempty"`
	LastScheduleTime      *metav1.Time         `json:"lastScheduleTime,omitempty"`
	LastScheduleMicroTime *metav1.MicroTime    `json:"lastScheduleMicroTime,omitempty"`
}

// CronJobStatusApplyConfiguration constructs an declarative configuration of the CronJobStatus type for use with
// apply.
func CronJobStatus() *CronJobStatusApplyConfiguration {
	return &CronJobStatusApplyConfiguration{}
}

// WithActive adds the given value to the Active field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Active field.
func (b *CronJobStatusApplyConfiguration) WithActive(values ...v1.ObjectReference) *CronJobStatusApplyConfiguration {
	for i := range values {
		b.Active = append(b.Active, values[i])
	}
	return b
}

// WithLastScheduleTime sets the LastScheduleTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastScheduleTime field is set to the value of the last call.
func (b *CronJobStatusApplyConfiguration) WithLastScheduleTime(value metav1.Time) *CronJobStatusApplyConfiguration {
	b.LastScheduleTime = &value
	return b
}

// WithLastScheduleMicroTime sets the LastScheduleMicroTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastScheduleMicroTime field is set to the value of the last call.
func (b *CronJobStatusApplyConfiguration) WithLastScheduleMicroTime(value metav1.MicroTime) *CronJobStatusApplyConfiguration {
	b.LastScheduleMicroTime = &value
	return b
}
