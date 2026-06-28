/*
Copyright 2026 The Kubernetes Authors.

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
package emptyobjectdefault

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyAuditConfig is a nested struct whose fields each define a default.
type PolicyAuditConfig struct {
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=1
	// +optional
	RateLimit *uint32 `json:"rateLimit,omitempty"`

	// +kubebuilder:default=50
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxFileSize *uint32 `json:"maxFileSize,omitempty"`

	// +kubebuilder:default=null
	// +optional
	Destination string `json:"destination,omitempty"`

	// +kubebuilder:default=local0
	// +optional
	SyslogFacility string `json:"syslogFacility,omitempty"`
}

// EmptyObjectDefaultSpec defaults a pointer field to an empty object. The nested
// fields still need their defaults, so the parent must render as default: {} and
// not default: null.
type EmptyObjectDefaultSpec struct {
	// +kubebuilder:default={}
	// +optional
	PolicyAuditConfig *PolicyAuditConfig `json:"policyAuditConfig,omitempty"`
}

// +kubebuilder:object:root=true

// EmptyObjectDefault is the Schema for the emptyobjectdefault API
type EmptyObjectDefault struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EmptyObjectDefaultSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// EmptyObjectDefaultList contains a list of EmptyObjectDefault
type EmptyObjectDefaultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EmptyObjectDefault `json:"items"`
}

// StageConfig has a required field. When the parent defaults to an empty object,
// the API server rejects the CRD because the default must satisfy the schema and
// the required field is missing.
type StageConfig struct {
	// +kubebuilder:validation:Enum=pending;running;done
	Stage string `json:"stage"`
}

// RequiredChildDefaultSpec defaults a pointer field to an empty object whose
// child is required.
type RequiredChildDefaultSpec struct {
	// +kubebuilder:default={}
	// +optional
	Status *StageConfig `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RequiredChildDefault is the Schema for the requiredchilddefaults API
type RequiredChildDefault struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RequiredChildDefaultSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// RequiredChildDefaultList contains a list of RequiredChildDefault
type RequiredChildDefaultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RequiredChildDefault `json:"items"`
}
