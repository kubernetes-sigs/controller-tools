/*
Copyright 2019 The Kubernetes Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// ExampleV1 is a kind with schema changes intended to be generated as
// a V1 CRD.
type ExampleV1 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +required
	Spec ExampleV1Spec `json:"spec"`
}

type ExampleV1Spec struct {
	// foo contains foo.
	Foo string `json:"foo"`
	// foo contains foo.
	Bar string `json:"bar"`
}

// +kubebuilder:object:root=true

// ExampleV1List contains a list of ExampleV1.
type ExampleV1List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ExampleV1 `json:"items"`
}
