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
package known_types

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// KnownTypeSpec uses the types of the packages that have known-schema overrides:
// Quantity and IntOrString are overridden, while Format and Type are taken from
// the package itself.
type KnownTypeSpec struct {
	Size resource.Quantity `json:"size"`

	// How Size is printed back: DecimalSI or BinarySI.
	SizeFormat resource.Format `json:"sizeFormat,omitempty"`

	Target intstr.IntOrString `json:"target"`

	// Whether Target holds an integer or a string.
	TargetType intstr.Type `json:"targetType,omitempty"`
}

// +kubebuilder:object:root=true

// KnownType is the Schema for the knowntypes API
type KnownType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KnownTypeSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// KnownTypeList contains a list of KnownType
type KnownTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KnownType `json:"items"`
}
