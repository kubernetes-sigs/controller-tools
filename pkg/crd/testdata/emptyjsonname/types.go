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

//go:generate ../../../../.run-controller-gen.sh crd paths=./... output:dir=..

// +groupName=testdata.kubebuilder.io
// +versionName=v1
package emptyjsonname

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RetentionPolicy is nested under the fields that have an empty name in their
// JSON tag.
type RetentionPolicy struct {
	Days int32 `json:"days,omitempty"`
}

// InlineRetention is embedded; encoding/json inlines it regardless of its tag.
type InlineRetention struct {
	Mode string `json:"mode,omitempty"`
}

// EmptyJsonNameSpec checks that a named field with an empty JSON name keeps the
// Go field name, while an embedded field with an empty JSON name stays inline.
type EmptyJsonNameSpec struct {
	Schedule string `json:"schedule"`

	// omitempty alone does not make a named field anonymous.
	Retention RetentionPolicy `json:",omitempty"`

	// An explicitly empty name behaves the same way.
	Archive RetentionPolicy `json:""`

	InlineRetention `json:",omitempty"`
}

// +kubebuilder:object:root=true

// EmptyJsonName is the Schema for the emptyjsonnames API
type EmptyJsonName struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EmptyJsonNameSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// EmptyJsonNameList contains a list of EmptyJsonName
type EmptyJsonNameList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EmptyJsonName `json:"items"`
}
